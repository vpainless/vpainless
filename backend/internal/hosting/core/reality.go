package core

import (
	"crypto/rand"
	"encoding/base64"
	"encoding/hex"
	"fmt"
	"net"

	"github.com/gofrs/uuid/v5"
	"golang.org/x/crypto/curve25519"
)

type XrayConfig struct {
	ConnectionString string
}

type XrayTemplateID struct{ uuid.UUID }

type XrayTemplate struct {
	ID                   XrayTemplateID
	FakeURL              string
	Curve25519PrivateKey string
	Curve25519PublicKey  string
	ShortID              string
	Base                 string
}

const defaultXrayTemplate = `
{
  "inbounds": [
    {
      "listen": "0.0.0.0",
      "port": 443,
      "protocol": "vless",
      "settings": {
        "clients": [
          {
            "id": "%[1]s",
            "flow": "xtls-rprx-vision"
          }
        ],
        "decryption": "none"
      },
      "streamSettings": {
        "network": "tcp",
        "security": "reality",
        "realitySettings": {
          "show": false,
          "dest": "%[2]s:443",
          "xver": 0,
          "serverNames": ["%[2]s"],
          "privateKey": "%[3]s",
          "minClientVer": "1.8.0",
          "maxClientVer": "",
          "maxTimeDiff": 0,
          "shortIds": ["%[4]s"]
        }
      },
      "sniffing": {
        "enabled": true,
        "destOverride": ["http", "tls", "quic"]
      }
    }
  ],
  "routing": {
    "domainStrategy": "IPIfNonMatch",
    "rules": [
      {
        "type": "field",
        "outboundTag": "block",
        "ip": [
          "geoip:ir",
          "geoip:private",
          "192.168.0.0/16",
          "10.0.0.0/8",
          "172.16.0.0/12",
          "127.0.0.0/8"
        ]
      },
      {
        "type": "field",
        "outboundTag": "block",
        "domain": [
          "geosite:private",
          "geosite:category-ir",
          "snapp",
          "digikala",
          "tapsi",
          "blogfa",
          "bank",
          "sb24.com",
          "sheypoor.com",
          "tebyan.net",
          "beytoote.com",
          "telewebion.com",
          "Film2movie.ws",
          "Setare.com",
          "Filimo.com",
          "Torob.com",
          "Tgju.org",
          "Sarzamindownload.com",
          "downloadha.com",
          "P30download.com",
          "Sanjesh.org",
          "domain:intrack.ir",
          "domain:divar.ir",
          "domain:irancell.ir",
          "domain:yooz.ir",
          "domain:iran-cell.com",
          "domain:irancell.i-r",
          "domain:shaparak.ir",
          "domain:learnit.ir",
          "domain:yooz.ir",
          "domain:baadesaba.ir",
          "domain:webgozar.ir",
          "domain:dt.beyla.site"
        ]
      }
    ]
  },
  "outbounds": [
    {
      "protocol": "freedom",
      "tag": "direct"
    },
    {
      "protocol": "blackhole",
      "tag": "block"
    }
  ],
  "log": {
    "loglevel": "warning"
  },
  "policy": {
    "levels": {
      "0": {
        "handshake": 3,
        "connIdle": 180
      }
    }
  }
}
`

// NewRealityConfig creates a new reality config.
//
// fakeURL should not include the protocol schema, and should be global geosite
// accessible from within the country we would like to bypass the censorship.
// e.g. "www.speedtest.net"
func NewRealityConfig(fakeURL string) (XrayTemplate, error) {
	prv, pub, err := genCurve25519KeyPair()
	if err != nil {
		return XrayTemplate{}, fmt.Errorf("error generating curve 25519 key pair: %w", err)
	}

	shortID, err := genShortID()
	if err != nil {
		return XrayTemplate{}, fmt.Errorf("error generating short id: %w", err)
	}

	return XrayTemplate{
		ID:                   XrayTemplateID{uuid.Must(uuid.NewV4())},
		FakeURL:              fakeURL,
		Curve25519PrivateKey: base64.RawURLEncoding.EncodeToString(prv),
		Curve25519PublicKey:  base64.RawURLEncoding.EncodeToString(pub),
		ShortID:              hex.EncodeToString(shortID),
		Base:                 defaultXrayTemplate,
	}, nil
}

func (c XrayTemplate) String() string {
	return fmt.Sprintf(c.Base, c.ID, c.FakeURL, c.Curve25519PrivateKey, c.ShortID)
}

// ConnectionString creates a connection string to connect to a server made
// by this reality config
func (c XrayTemplate) ConnectionString(ipv4 net.IP) string {
	format := `vless://%s@%s:443?flow=xtls-rprx-vision&type=raw&security=reality&sni=%s&pbk=%s&sid=%s#xray`
	return fmt.Sprintf(format, c.ID, ipv4.String(), c.FakeURL, c.Curve25519PublicKey, c.ShortID)
}

func genCurve25519KeyPair() (privateKey, publicKey []byte, err error) {
	privateKey = make([]byte, curve25519.ScalarSize)
	if _, err = rand.Read(privateKey); err != nil {
		return nil, nil, err
	}

	// Modify random bytes using algorithm described at:
	// https://cr.yp.to/ecdh.html.
	privateKey[0] &= 248
	privateKey[31] &= 127
	privateKey[31] |= 64

	if publicKey, err = curve25519.X25519(privateKey, curve25519.Basepoint); err != nil {
		return nil, nil, err
	}

	return
}

func genShortID() ([]byte, error) {
	id := make([]byte, 3)
	if _, err := rand.Read(id); err != nil {
		return nil, err
	}

	return id, nil
}
