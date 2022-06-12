package token

import (
	"testing"
	"time"

	"github.com/dgrijalva/jwt-go"
)

const privteKey = `-----BEGIN RSA PRIVATE KEY-----
MIIEowIBAAKCAQEAzlUMe+vM2qDMf3bHyHQXwSnrEurI4QO9O3i9gRojjEmWZegw
o9f2vrXejeaTrNM+WjHHjpoTJIFTNARPdi+tBCAYPRJTKR/QmaUOb+48DQoEVCkR
8WDGApn4R+SFNAjaApUeJMvDeP2qvFFbJ0A0YKMF0iHIfMZEfOsXsi44deYhVlHg
uKXowQ+A/Zo4CmKwW9Y5cXwgAiqqp/wA7LlTeemF+cYO+sLXCexZ/vczBrTg5Z+A
MjjYf/WNFqm5Tzik8o//ckg6vE0SuwKzpXpI62vEcUuUfrAp14+vMB3XHk1fY0id
IWONb6Hg2STccaXIXVwEjP9s6HKv8vr7AR+RfQIDAQABAoIBAQCrEvUw4geN1fj4
TkHDQA5aCClyG9zGRFVns+pb2pJSxMjAYc3Ca1OYOC74tI8IonV2TwPIhpMMl3Wn
EVPZCBqJ6xptuH3fARPx8FqSD5MWtJF2Pj80RSqoCYVEBoMy64vmzECb/Z5q+NfR
IBtV5fQHk+NFoOEcIz+x2zJgd9Y0WUbNjdX4XD4ogrSdJnSbqeyZLo0yy8UizmF6
SAjIpxdedYyJuwrhscLDbUE8SJFqMDEnLxSsbunaQaBw7FEX/Yz9ZB2pmzxN4oVS
hULORYhI6txHyvkDfcZ9UL064dz1WswVaxjVsZDMHq0538z8vU3QB8bk9r2jfLOG
bj6Tw+X5AoGBAPPd+tkoDxgk0/jvwbfn70SporUCCruj/zERVTBGbR1eEKmMwmG5
85kqgh3YXkIl6CQQ/P144T64tGO90vxeZ1MO/3ycaB/81e4P3KyeK7afXYRCDP/F
VOhUC10sZBOWwj0shDyYZheq5otLnWDFMdLqiohptt8hzjJZ07mwdwcjAoGBANiZ
ATiW02iXesK1uK6PBhvmsEdmQF36B5yPy1m5lg4wEEBRM3vRuYB3ItP/IKLFfFt6
8N5tFBF575jXRpyBygzuPCfmcgMq4zHfvgGVpEQ6wo3Z5iDL+yV+2NGdqlSaHcwt
ENbqbSIbiMuCtRV+VBRxN+mfsyqmb1CNqLJdwN7fAoGAS56PxHq5g4EYAd9GsKJI
/X+kpoBFl73YyfxX8CpGd47Nl+W/+NHSibI2us53HAfpHhXufSLYpbxco3kfTYZw
f77s1lUhrJmYNMPSZ+x8HZr5QqPAqCcmlwxIodG8Dp73CEUflDKlpb0m1BbUbEd5
la+I2Zf+Tt6Ks+5Kyw+/OQUCgYAS1cNQa3U43CtVsT29GDzcwkPEAbVJNsvgpnR1
efj9hNp07Vq/wq4R6MpDiyUIYon83oUBopSjLGpUbSv7wiGS3Eio45Y4hks5dA5u
ztd5A28VrMQhR/uv+Abcu4wrTTeYILcdKUeSNri/kb8zfkfLe0j0bOnEpLJ7W8Y2
tIZoGwKBgCaronugacmnVa0aj7emiz0APOLuhPSEqnpH9mtmcCrsfydm23IyDBCn
TFNxSj5ZfkR7Uuw5mKSveqHLfKU046VhbnH/BcubZ3zf5ravf/m49c3aTwl6Hsmj
lLJslfcXswCHcHZTvtaSZCw2staXRT1sOCrO33DNj+V0Y2niEdcF
-----END RSA PRIVATE KEY-----`

func TestGenerateToken(t *testing.T) {
	key, err := jwt.ParseRSAPrivateKeyFromPEM([]byte(privteKey))
	if err != nil {
		t.Fatalf("cannot parse private key: %v", err)
	}
	g := NewJWTTokenGen("coolcar/auth", key)
	g.nowFunc = func() time.Time {
		return time.Unix(1516239022, 0)
	}
	token, err := g.GenerateToken("629ff7b89701e5e6ea719e4b", 2*time.Hour)
	if err != nil {
		t.Errorf("cannot generate token: %v", err)
	}

	want := "eyJhbGciOiJSUzUxMiIsInR5cCI6IkpXVCJ9.eyJleHAiOjE1MTYyNDYyMjIsImlhdCI6MTUxNjIzOTAyMiwiaXNzIjoiY29vbGNhci9hdXRoIiwic3ViIjoiNjI5ZmY3Yjg5NzAxZTVlNmVhNzE5ZTRiIn0.cv4PQonR17bJ-MiR1zLmS8nOIcCAt8IiQIY8QBwc8pSPzsM9rIpK3nSuygJmSQ7DTXlHjMxjJaxo3R0DbvMZxDZbqJnTQBk37ogwTnWO3OPOx_atrURycBDkCXdo4v6YaGs2JqNdfy_IygsM5h4SMXy6e_l0Gx4Ybf5yBXTKO-l0ZPsLYG96gEhGY7S_E5-3rb6TgUu3V2u6GAXb81ZcIxW19jH15PTsXf8CUQaavsl32anV1CXIgQGfT8dONF96YN-hrARoi7DL4BTQSc2LPt2Beyp4KcnEiKlMdGDYAP5827loY7aABx7sq4E_X63-lD4wAG6mDAZrr-fT1P4wNg"
	if token != want {
		t.Errorf("wrong token generated. want: %q; got: %q", want, token)
	}
}
