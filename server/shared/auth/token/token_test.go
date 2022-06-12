package token

import (
	"testing"
	"time"

	"github.com/dgrijalva/jwt-go"
)

const publicKey = `-----BEGIN PUBLIC KEY-----
MIIBIjANBgkqhkiG9w0BAQEFAAOCAQ8AMIIBCgKCAQEAzlUMe+vM2qDMf3bHyHQX
wSnrEurI4QO9O3i9gRojjEmWZegwo9f2vrXejeaTrNM+WjHHjpoTJIFTNARPdi+t
BCAYPRJTKR/QmaUOb+48DQoEVCkR8WDGApn4R+SFNAjaApUeJMvDeP2qvFFbJ0A0
YKMF0iHIfMZEfOsXsi44deYhVlHguKXowQ+A/Zo4CmKwW9Y5cXwgAiqqp/wA7LlT
eemF+cYO+sLXCexZ/vczBrTg5Z+AMjjYf/WNFqm5Tzik8o//ckg6vE0SuwKzpXpI
62vEcUuUfrAp14+vMB3XHk1fY0idIWONb6Hg2STccaXIXVwEjP9s6HKv8vr7AR+R
fQIDAQAB
-----END PUBLIC KEY-----`

func TestVerify(t *testing.T) {

	pubKey, err := jwt.ParseRSAPublicKeyFromPEM([]byte(publicKey))
	if err != nil {
		t.Fatalf("cannot parse public key: %v", err)
	}

	v := &JWTTokenVerifier{
		PublicKey: pubKey,
	}

	cases := []struct {
		name    string
		token   string
		now     time.Time
		want    string
		wantErr bool
	}{
		{
			name:  "valid_token",
			token: "eyJhbGciOiJSUzUxMiIsInR5cCI6IkpXVCJ9.eyJleHAiOjE1MTYyNDYyMjIsImlhdCI6MTUxNjIzOTAyMiwiaXNzIjoiY29vbGNhci9hdXRoIiwic3ViIjoiNjI5ZmY3Yjg5NzAxZTVlNmVhNzE5ZTRiIn0.cv4PQonR17bJ-MiR1zLmS8nOIcCAt8IiQIY8QBwc8pSPzsM9rIpK3nSuygJmSQ7DTXlHjMxjJaxo3R0DbvMZxDZbqJnTQBk37ogwTnWO3OPOx_atrURycBDkCXdo4v6YaGs2JqNdfy_IygsM5h4SMXy6e_l0Gx4Ybf5yBXTKO-l0ZPsLYG96gEhGY7S_E5-3rb6TgUu3V2u6GAXb81ZcIxW19jH15PTsXf8CUQaavsl32anV1CXIgQGfT8dONF96YN-hrARoi7DL4BTQSc2LPt2Beyp4KcnEiKlMdGDYAP5827loY7aABx7sq4E_X63-lD4wAG6mDAZrr-fT1P4wNg",
			now:   time.Unix(1516239122, 0),
			want:  "629ff7b89701e5e6ea719e4b",
		},
		{
			name:    "token_expired",
			token:   "eyJhbGciOiJSUzUxMiIsInR5cCI6IkpXVCJ9.eyJleHAiOjE1MTYyNDYyMjIsImlhdCI6MTUxNjIzOTAyMiwiaXNzIjoiY29vbGNhci9hdXRoIiwic3ViIjoiNjI5ZmY3Yjg5NzAxZTVlNmVhNzE5ZTRiIn0.cv4PQonR17bJ-MiR1zLmS8nOIcCAt8IiQIY8QBwc8pSPzsM9rIpK3nSuygJmSQ7DTXlHjMxjJaxo3R0DbvMZxDZbqJnTQBk37ogwTnWO3OPOx_atrURycBDkCXdo4v6YaGs2JqNdfy_IygsM5h4SMXy6e_l0Gx4Ybf5yBXTKO-l0ZPsLYG96gEhGY7S_E5-3rb6TgUu3V2u6GAXb81ZcIxW19jH15PTsXf8CUQaavsl32anV1CXIgQGfT8dONF96YN-hrARoi7DL4BTQSc2LPt2Beyp4KcnEiKlMdGDYAP5827loY7aABx7sq4E_X63-lD4wAG6mDAZrr-fT1P4wNg",
			now:     time.Unix(1517239122, 0),
			wantErr: true,
		},
		{
			name:    "bad_token",
			token:   "bad_token",
			now:     time.Unix(1517239122, 0),
			wantErr: true,
		},
		{
			name:    "wrong_signature",
			token:   "eyJhbGciOiJSUzUxMiIsInR5cCI6IkpXVCJ9.eyJleHAiOjE1MTYyNDYyMjIsImlhdCI6MTUxNjIzOTAyMiwiaXNzIjoiY29vbGNhci9hdXRoIiwic3ViIjoiNWY3YzMxNjhlMjI4M2FhNzIyZTM1MWE0In0.jPVRIZXsNz08OCudP4cC8KGzVEIWC42TOMHpc6cN-_3yUgbPcrhuJL6C27fzoxt0j8J3L0z6nv0ni_713fzYjo1Y_b4Axxz4sI5bz-b9O1BziFU1NC9t3IJbwFsF2Svz2OpG3aY388rTZ4orHShfRbrzGnzK8NbNXIZ7CcCvEznHiJEmSgqSZSYeZVjjid2p2l_T_eTQxJTkHi9LE-3g_AfLKLXXmqLlXYpurTGMWEBkJq51uNs6MnESi4pEwbLviTmZTTtC6qAhkVmeJh7QUZA8BPKoxSbNEYQxYYQK1aiRGyrrONsK1etXW6JG2F4x0wiNjTKMvQSAsq7GnWvkoQ",
			now:     time.Unix(1516239122, 0),
			wantErr: true,
		},
	}

	for _, c := range cases {
		t.Run(c.name, func(t *testing.T) {
			jwt.TimeFunc = func() time.Time {
				return c.now
			}
			accountID, err := v.Verify(c.token)

			if !c.wantErr && err != nil {
				t.Errorf("verification failed: %v", err)
			}

			if c.wantErr && err == nil {
				t.Errorf("want error; got no error")
			}

			if accountID != c.want {
				t.Errorf("wrong account id. want: %q, got: %q", c.want, accountID)
			}
		})
	}
}
