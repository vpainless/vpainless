package authz

import (
	"context"
	"encoding/json"
	"fmt"

	"github.com/gofrs/uuid/v5"
	"github.com/open-policy-agent/opa/v1/rego"
)

// Clauses are additional conditions passed to the DB
// to enforce authz policies.
type Clause struct {
	Condition string `json:"condition"`
	Values    []any  `json:"values"`
}

// IsNil is used to determine if a clause should be
// applied on the query or not.
func (c *Clause) IsNil() bool {
	return c.Condition == ""
}

// Policy encapsulates the outcome of a policy evaluations.
type Policy struct {
	// When false, the principal is not authorized to perform
	// the requested verb on the specified resource.
	Allow bool `json:"allow"`
	// Partial is an additional clause to be passed to the DB
	// for further filtration. Partial should be discarded when
	// Allow is false.
	Partial Clause `json:"partial"`
}

type input struct {
	Principal map[string]any `json:"principal,omitempty"`
	Operation Verb           `json:"action"`
	Resource  any            `json:"resource,omitempty"`
	// This is not required. It's just to make it eaiser to debug policies.
	// Resources are evaluated agains policies packaged for their group.
	// For example, if the resource group is "users", we refer to
	// access.users rego package.
	ResourceGroup string `json:"resource_group,omitempty"`
}

func (i input) String() string {
	b, err := json.MarshalIndent(i, "", "  ")
	if err != nil {
		return ""
	}

	return string(b)
}

// Validator validates users requests against a set of pre defined rego policies
// using OPA(open policy agent). The result is a Policy that should be enforced
// later by the core application. This will be performed by disallowing users
// to perform requested actions, or allowing partial execution of the actions
// by passing additional DB clauses.
type Validator struct {
	defaultOptions []func(*rego.Rego)
	pkgname        string
}

type ValidatorOption func(e *Validator)

func WithRegoModule(path, content string) ValidatorOption {
	return func(v *Validator) {
		v.defaultOptions = append(v.defaultOptions, rego.Module(path, content))
	}
}

// NewValidator creates a new authz validator
func NewValidator(pkgname string, opts ...ValidatorOption) *Validator {
	v := &Validator{pkgname: pkgname}

	for _, opt := range opts {
		opt(v)
	}

	return v
}

// Can verifies if the principal can perform the requested verb on the specified resource.
func (v *Validator) Can(ctx context.Context, p Principal, op Verb, resourcer Resourcer) (Policy, error) {
	rgroup, rvalue := resourcer.Resource()

	input := input{
		Operation: op,
		Resource:  rvalue,
		// Resource group is not needed in the input,
		// as it is imposed by the policy module we run
		// the validation against. We just include it
		// for debugging purposes.
		ResourceGroup: rgroup,
	}

	if !p.ID.IsNil() {
		input.Principal = map[string]any{
			"id":   p.ID,
			"role": p.Role,
		}
	}

	if p.GroupID != uuid.Nil {
		input.Principal["group_id"] = p.GroupID
	}

	// TODO: remove this
	fmt.Println(input)

	options := append(v.defaultOptions,
		rego.Query(fmt.Sprintf(`policy := {"allow": data.%[1]s.%[2]s.allow, "partial": data.%[1]s.%[2]s.partial}`, v.pkgname, rgroup)),
	)
	query := rego.New(options...)

	prepared, err := query.PrepareForEval(ctx)
	if err != nil {
		return Policy{}, err
	}

	result, err := prepared.Eval(ctx, rego.EvalInput(input))
	if err != nil {
		return Policy{}, err
	}

	if len(result) == 0 {
		return Policy{}, fmt.Errorf("unexpected empty policy result")
	}

	b, err := json.Marshal(result[0].Bindings["policy"])
	if err != nil {
		return Policy{}, err
	}

	var policy Policy
	if err = json.Unmarshal(b, &policy); err != nil {
		return Policy{}, err
	}
	return policy, nil
}
