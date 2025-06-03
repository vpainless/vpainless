package storage

import (
	"context"
	"database/sql"
	"errors"
	"log/slog"
	"net/url"

	"vpainless/internal/hosting/core"
	"vpainless/internal/pkg/db"
	"vpainless/pkg/querybuilder"

	"github.com/gofrs/uuid/v5"
)

func (r *Repository) GetGroup(ctx context.Context, id core.GroupID) (*core.Group, error) {
	var group *core.Group
	if err := r.db.InTxDo(ctx, sql.LevelReadCommitted, func(ctx context.Context, tx *db.Tx) error {
		var err error

		q := groupGetQuery{id}
		query, args := q.SQL()
		group, err = scanGroup(tx.QueryRowContext(ctx, query, args...))
		if err != nil {
			return err
		}

		xq := xrayTemplateGetQuery{id}
		query, args = xq.SQL()
		rows, err := tx.QueryContext(ctx, query, args...)
		if err != nil {
			return err
		}
		templates, err := scanXrayTemplates(rows)
		if err != nil {
			return err
		}

		group.XrayTemplates = make(map[core.XrayTemplateID]core.XrayTemplate, len(templates))
		for _, t := range templates {
			group.XrayTemplates[t.ID] = t
		}

		sq := sshKeyPairGetQuery{group.DefaultSSHKey.ID}
		query, args = sq.SQL()
		row := tx.QueryRowContext(ctx, query, args...)
		key, err := scanSSHKeyPair(row)
		if err != nil {
			return err
		}
		group.DefaultSSHKey = key

		ssq := startupScriptGetQuery{group.DefaultStartUpScript.ID}
		query, args = ssq.SQL()
		row = tx.QueryRowContext(ctx, query, args...)
		script, err := scanStartupScript(row)
		if err != nil {
			return err
		}
		group.DefaultStartUpScript = script

		return nil
	}); err != nil {
		return nil, err
	}

	return group, nil
}

type groupGetQuery struct {
	groupID core.GroupID
}

func (q groupGetQuery) SQL() (string, []any) {
	qb := querybuilder.New(`
		select
			g.id, g.name, g.provider_name, g.provider_url, g.provider_apikey, g.default_xray_template, g.default_ssh_key, g.default_startup_script
		from groups g
		where g.id = ?`, q.groupID,
	)
	return qb.SQL()
}

func scanGroup(row *sql.Row) (*core.Group, error) {
	var group core.Group
	var u string
	if err := row.Scan(
		&group.ID,
		&group.Name,
		&group.Host.Name,
		&u,
		&group.Host.APIKey,
		&group.DefaultXrayTemplate,
		&group.DefaultSSHKey.ID,
		&group.DefaultStartUpScript.ID,
	); err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return nil, core.ErrNotFound
		}
		return nil, err
	}

	parsed, err := url.Parse(u)
	if err != nil {
		return nil, err
	}

	group.Host.Base = *parsed
	return &group, nil
}

type xrayTemplateGetQuery struct {
	groupID core.GroupID
}

func (q xrayTemplateGetQuery) SQL() (string, []any) {
	qb := querybuilder.New(`
			select id, base
			from xray_templates
			where group_id = ?;`, q.groupID,
	)
	return qb.SQL()
}

func scanXrayTemplates(rows *sql.Rows) ([]core.XrayTemplate, error) {
	var result []core.XrayTemplate
	for rows.Next() {
		var t core.XrayTemplate
		if err := rows.Scan(&t.ID, &t.Base); err != nil {
			return nil, err
		}

		result = append(result, t)
	}
	if err := rows.Err(); err != nil {
		return nil, err
	}

	return result, nil
}

type sshKeyPairGetQuery struct {
	id core.SSHKeyID
}

func (q sshKeyPairGetQuery) SQL() (string, []any) {
	qb := querybuilder.New(`
			select id, remote_id, name, private_key, public_key
			from ssh_keys
			where id = ?;`, q.id,
	)
	return qb.SQL()
}

func scanSSHKeyPair(row *sql.Row) (core.SSHKeyPair, error) {
	var result core.SSHKeyPair
	var remoteID sql.NullString

	if err := row.Scan(&result.ID, &remoteID, &result.Name, &result.PrivateKey, &result.PublicKey); err != nil {
		return core.SSHKeyPair{}, err
	}

	if remoteID.Valid {
		result.RemoteID = core.SSHKeyID{UUID: uuid.FromStringOrNil(remoteID.String)}
	} else {
		result.RemoteID = core.SSHKeyID{UUID: uuid.Nil}
	}

	return result, nil
}

type startupScriptGetQuery struct {
	id core.StartUpScriptID
}

func (q startupScriptGetQuery) SQL() (string, []any) {
	qb := querybuilder.New(`
			select id, remote_id, content
			from startup_scripts
			where id = ?;`, q.id,
	)
	return qb.SQL()
}

func scanStartupScript(row *sql.Row) (core.StartUpScript, error) {
	var result core.StartUpScript
	var remoteID sql.NullString

	if err := row.Scan(&result.ID, &remoteID, &result.Content); err != nil {
		return core.StartUpScript{}, err
	}

	if remoteID.Valid {
		result.RemoteID = core.StartUpScriptID{UUID: uuid.FromStringOrNil(remoteID.String)}
	} else {
		result.RemoteID = core.StartUpScriptID{UUID: uuid.Nil}
	}

	return result, nil
}

func (r *Repository) SaveGroup(ctx context.Context, group *core.Group) (*core.Group, error) {
	logger := slog.With("group_id", group.ID)
	if err := r.db.InTxDo(ctx, sql.LevelLinearizable, func(ctx context.Context, tx *db.Tx) error {
		logger.InfoContext(ctx, "DB: insert xray templates...")
		xrayquery := `
			insert into xray_templates (id, base, group_id) values (?, ?, ?)
			on conflict do nothing;
		`
		prepared, err := tx.PrepareContext(ctx, xrayquery)
		if err != nil {
			return err
		}
		for _, t := range group.XrayTemplates {
			_, err := prepared.ExecContext(ctx, t.ID, t.Base, group.ID)
			if err != nil {
				return err
			}
		}

		logger.InfoContext(ctx, "DB: insert ssh key...")
		sshquery := `
			insert into ssh_keys (id, group_id, remote_id, name, private_key, public_key)
			values (?, ?, ?, ?, ?, ?)
			on conflict do update set
				remote_id = excluded.remote_id
			;
		`
		_, err = tx.ExecContext(ctx, sshquery,
			group.DefaultSSHKey.ID,
			group.ID,
			uuidOrNull(group.DefaultSSHKey.RemoteID.UUID),
			group.DefaultSSHKey.Name,
			group.DefaultSSHKey.PrivateKey,
			group.DefaultSSHKey.PublicKey,
		)
		if err != nil {
			return err
		}

		logger.InfoContext(ctx, "DB: insert startup script...")
		scriptquery := `
			insert into startup_scripts (id, group_id, remote_id, content)
			values (?, ?, ?, ?)
			on conflict do update set
				remote_id = excluded.remote_id
			;
		`
		_, err = tx.ExecContext(ctx, scriptquery,
			group.DefaultStartUpScript.ID,
			group.ID,
			uuidOrNull(group.DefaultStartUpScript.RemoteID.UUID),
			group.DefaultStartUpScript.Content,
		)
		if err != nil {
			return err
		}

		logger.InfoContext(ctx, "DB: insert group...")
		qb := querybuilder.New(`
			insert into groups (
				id,
				name,
				provider_name,
				provider_url,
				provider_apikey,
				default_xray_template,
				default_ssh_key,
				default_startup_script
			)
			values (?, ?, ?, ?, ?, ?, ?, ?)
			on conflict (id) do update set
				name = excluded.name,
				provider_name = excluded.provider_name,
				provider_url = excluded.provider_url,
				provider_apikey = excluded.provider_apikey,
				default_xray_template = excluded.default_xray_template,
				default_ssh_key = excluded.default_ssh_key,
				default_startup_script = excluded.default_startup_script;
		`,
			group.ID, group.Name, group.Host.Name, group.Host.Base.String(),
			group.Host.APIKey, group.DefaultXrayTemplate, group.DefaultSSHKey.ID,
			group.DefaultStartUpScript.ID,
		)

		query, args := qb.SQL()
		_, err = tx.ExecContext(ctx, query, args...)
		return err
	}); err != nil {
		return nil, err
	}

	return group, nil
}
