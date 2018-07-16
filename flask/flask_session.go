// https://gist.github.com/babldev/502364a3f7c9bafaa6db
// https://windard.com/blog/2017/10/17/Flask-Session
// session value = [是否压缩(zlib)标志.]+ base64Endcode(Data) + "." + base64Endcode(Timestamp) + "." + base64Endcode(Signature)
// flask 无密码登录
// flask session cookie decoder / encoder
package main

import (
	"bytes"
	"compress/zlib"
	"crypto/hmac"
	"crypto/rand"
	"crypto/sha1"
	"crypto/sha512"
	"encoding/base64"
	"encoding/hex"
	"errors"
	"fmt"
	"hash"
	"io"
	"net/http"
	"strings"
	"time"

	"github.com/gin-gonic/gin"
)

var (
	// 2011/01/01 in UTC from python package: itsdangerous
	EPOCH    int64 = 1293840000
	base64er       = base64.RawURLEncoding
)

// func main() {
// 	// session := ".eJw9kE1rg0AURf9KmXUWzhg3QhbKs2LgvcGiCW82gaqJjp0WTIIfIf-9NouuL_dw7n2I03lorq0Ib8O92YhTV4vwId4-RSioSAKC3DMu8XWBI6vMQ2U6kx6-sIhGgrhDh4oUtVyU0qTZiDaSbOuO7GXhJZ9xuQSU8qghn_iYKbbltGbKQBWwymeT4oTQK7SJRBs7hLpjhz6rgzXHXJJF30C_UIGzLvotqeSvHyBEPtu9I1t5-ljuxHMjqutwPt1--ub7f4JJ962GbNJQ-pTmUkNr2SWr2nvLKvEMfKzqcU_QBytmZldKHncv3P3aDK87xFaK5y-RXl9G.Di2Bdw.K5r2N_KEfgd7NYIhip7H3spSYns"
// 	// //session := "eyJjc3JmX3Rva2VuIjp7IiBiIjoiWkdKaE9ESXhPRFUzTkdRMU9EaGpZbUV6WXpGaFkyRTBaRFJoWVRCa05EazVPV1V5WW1VMVl3PT0ifX0.DioKyw.2k2KhKSlSIU2vvNMNKsVWXlsU9g"
// 	// data, err := DecodeFlaskSession(session, "somexxxkey", "cookie-session")
// 	// fmt.Println(string(data), err)

// 	// rawData := fmt.Sprintf(`{"_fresh":true,"_id":{" b":"%s"},"csrf_token":{" b":"%s"},"user_id":"%d"}`,
// 	// 	base64URLEndcode(createIdentifier("127.0.0.1", "api")),
// 	// 	base64URLEndcode(generateCSRFToken()),
// 	// 	41,
// 	// )

// 	// fmt.Println("raw:", rawData)

// 	// sessionData, err := EncodeFlaskSession(rawData, "somexxxkey", "cookie-session")
// 	// fmt.Println("session:", sessionData, err)

// 	sessionData := "eyJjc3JmX3Rva2VuIjp7IiBiIjoiTmpoa05XSTRNak15TjJZd1ltWTJabUUwT0dSbFpqSXdZVGswTWpnMU5EWXpPVEkxWmpNNU5nPT0ifX0.Di21Ow.11l9E4Kz7cXf3HKebfyWBuGj3xw"
// 	rawData2, err := DecodeFlaskSession(sessionData, "somexxxkey", "cookie-session")
// 	fmt.Println(string(rawData2), err)
// }

func main() {
	//realSession := ".eJw9kE1rg0AQhv9K2XMOzSZehByEjYvCzGKxys5FaDRZ96MFTWg05L93yaHHYZ55P-bBuvM0zIal1-k2bFg39ix9sLcvljIUWYJrdodWJ6otFrIfFkQeSPQW6tLDigHE557a3FNoPMQ9imOcyesVfrE97pQsRxTVqlqMfLWQ1AkJE8iedlgbjzZ7B1lwbfuAvNoCR4PCLSCrBSUFCvoe-RE4JCC8g7WMNziqGh3I0lLbeG2LA3tu2Gmezt31xw3f_xW0JaOiVMScEsCVJEd1M2pbBuBFEm0dhNxEK6tDFaNUq74cXnK3eZhe72D7LXv-ARPsYcA.Di3IcQ.5_vaOe528k_TAGhualgWyx68obY"
	realSession := ".eJw9kMGOgkAQRH9lM2cPAvFi4sHNuASS7glmgExfTBZRaJjdBDVgG_990cOeK3lVrx7qcBrqS6PW1-FWL9ShPar1Q318q7UiLhqwnWBZeCf5HaQLyKe9k6_W6G4FYRIgp0wl9o67Ee2-Nbbxxu4iJ9mdOFk6dpGJsxXofCR7FtJpg3rP6PMA_C4yuhIM55xzcYyd42rlmDzZJCI9c6W6g-178kVvZj68umzaU1x4E8_7-OhJw0Y9F6q6DKfD9berf_4VTOkmKJMJZRsAZxN6CJH3Lcpn-9Kg0oUg2wjjZATJlkbvAjhv3rjbpR7ed6hQPf8A801ixg.Di3PGw.zIHylr9XU1_3-M3uVAGMgpGeqPs"
	rawData, err := DecodeFlaskSession(realSession, "somexxxkey", "cookie-session")
	fmt.Println(string(rawData), err)

	router := gin.Default()

	router.GET("/debug/reset", func(ctx *gin.Context) {
		ip, userAgent := ctx.ClientIP(), ctx.GetHeader("user-agent")
		fmt.Println(ip, userAgent)

		//ip = "221.12.20.22"
		//userAgent = "Mozilla/5.0 (X11; Linux x86_64) AppleWebKit/537.36 (KHTML, like Gecko) Chrome/63.0.3239.132 Safari/537.36"
		rawData := fmt.Sprintf(`{"_fresh":true,"_id":{" b":"%s"},"csrf_token":{" b":"%s"},"user_id":"%d"}`,
			base64URLEndcode(createIdentifier(ip, userAgent)), // _id 随 ip和userAgent变化,变化后会重新生成session
			base64URLEndcode(generateCSRFToken()),
			2,
		)

		fmt.Println(rawData)

		sessionData, err := EncodeFlaskSession(rawData, "somexxxkey", "cookie-session")
		fmt.Println("session:", sessionData, err)

		ctx.SetCookie("session", sessionData, 360000, "/", "", false, true)

		ctx.Redirect(http.StatusFound, "/")
	})

	router.Run(":8080")
}

//  "cookie-session" 是 flask session默认的salt
func DecodeFlaskSession(raw, secretKey, salt string) ([]byte, error) {
	var err error
	if raw, err = NewSigner(secretKey, salt, sha1.New).Load(raw); err != nil {
		return nil, err
	}

	var isCompress bool

	if strings.HasPrefix(raw, ".") {
		isCompress = true

		raw = raw[1:]
	}

	tmp := strings.Split(raw, ".")
	if len(tmp) != 2 {
		return nil, errors.New("Invalid Playload")
	}

	// handle Timestamp
	stamp := pareTimestamp(tmp[1])
	_ = stamp

	data := base64Decode(tmp[0])

	if isCompress {
		return zlibDecode(data), nil
	}

	return data, nil
}

func EncodeFlaskSession(raw, secretKey, salt string) (string, error) {
	data := fmt.Sprintf(".%s.%s",
		base64Endcode(zlibEncode([]byte(raw))),
		getTimestamp(),
	)

	return NewSigner(secretKey, salt, sha1.New).Dump(data)
}

type Signer struct {
	Secret string
	Salt   string
	Hash   func() hash.Hash
	Raw    string
	Sign   string
}

func NewSigner(secretKey, salt string, hash func() hash.Hash) *Signer {
	return &Signer{
		Secret: secretKey,
		Salt:   salt,
		Hash:   hash,
	}
}

func (s *Signer) Load(raw string) (string, error) {
	err := s.unsign(raw)
	if err != nil {
		return "", err
	}

	if !s.isTrue() {
		return "", errors.New("Invalid Sign")
	}

	return s.Raw, nil
}

func (s *Signer) Dump(raw string) (string, error) {
	s.Raw = raw

	err := s.sign(s.Raw)
	if err != nil {
		return "", err
	}

	return s.Raw + "." + s.Sign, nil
}

func (s *Signer) unsign(raw string) error {
	index := strings.LastIndex(raw, ".")
	if index == -1 {
		return errors.New("No Sign")
	}

	if index == len(raw)-1 {
		return errors.New("Empty Sign")
	}

	if index == 0 {
		return errors.New("Empty Playload")
	}

	s.Raw = raw[:index]
	s.Sign = raw[index+1:]

	return nil
}

func (s *Signer) sign(raw string) error {
	s.Sign = s.buildSign()

	return nil
}

func (s *Signer) isTrue() bool {
	return s.buildSign() == s.Sign
}

func (s *Signer) buildSign() string {
	mac := hmac.New(s.Hash, s.deriveKey())
	mac.Write([]byte(s.Raw))

	return base64Endcode(mac.Sum(nil))
}

func (s *Signer) deriveKey() []byte {
	mac := hmac.New(s.Hash, []byte(s.Secret))
	mac.Write([]byte(s.Salt))
	return mac.Sum(nil)
}

func zlibDecode(compressSrc []byte) []byte {
	b := bytes.NewReader(compressSrc)
	var out bytes.Buffer
	r, _ := zlib.NewReader(b)
	io.Copy(&out, r)
	return out.Bytes()
}

func zlibEncode(src []byte) []byte {
	var in bytes.Buffer
	w := zlib.NewWriter(&in)
	w.Write(src)
	w.Close()
	return in.Bytes()
}

func base64Endcode(src []byte) string {
	return base64er.EncodeToString(src)
}

func base64Decode(src string) []byte {
	dst := make([]byte, base64er.DecodedLen(len(src)))
	base64er.Decode(dst, []byte(src)) // signature sure no error

	return dst
}

func pareTimestamp(raw string) int64 {
	dst := base64Decode(raw)
	return bytes2int64(dst)
}

func getTimestamp() string {
	t := time.Now().Unix()

	return base64Endcode(int642bytes(t))
}

func bytes2int64(src []byte) int64 {
	var tmp int64
	n := len(src) - 1

	for i, v := range src {
		tmp |= int64(v) << byte(8*(n-i))
	}

	return tmp
}

func int642bytes(src int64) []byte {
	src -= EPOCH

	bs := make([]byte, 0)

	for src > 0 {
		bs = append(bs, byte(src&255))
		src >>= 8
	}

	return reversed(bs)
}

func reversed(bs []byte) []byte {
	l, r := 0, len(bs)-1

	for l < r {
		bs[l], bs[r] = bs[r], bs[l]
		l++
		r--
	}

	return bs
}

// from flask_login
func createIdentifier(remoteAddr, userAgent string) string {
	base := fmt.Sprintf("%s|%s", remoteAddr, userAgent)
	data := sha512.Sum512([]byte(base))

	return hex.EncodeToString(data[:])
}

func base64URLEndcode(src string) string {
	return base64.URLEncoding.EncodeToString([]byte(src))
}

func generateCSRFToken() string {
	b := make([]byte, 64)
	rand.Read(b)
	data := sha1.Sum(b)
	return hex.EncodeToString(data[:])
}

/*
# flask session decoder/encoder对应的python代码
from itsdangerous import URLSafeSerializer,URLSafeTimedSerializer
from flask.sessions import SecureCookieSessionInterface,TaggedJSONSerializer
import hashlib

def create_token():
	s = URLSafeSerializer('xxx')
	browser_id = 'ASDF'
	life_time = '100'
	token = s.dumps((1, 'admin', '123456', browser_id, life_time))
	return token

# https://gist.github.com/babldev/502364a3f7c9bafaa6db
def decode_flask_cookie(secret_key, cookie_str):
    salt = 'cookie-session'
    serializer = TaggedJSONSerializer()
    signer_kwargs = {
        'key_derivation': 'hmac',
        'digest_method': hashlib.sha1
    }
    s = URLSafeTimedSerializer(secret_key, salt=salt, serializer=serializer, signer_kwargs=signer_kwargs)
    return s.loads(cookie_str)

def encode_flask_cookie(secret_key, cookieDict):
            salt = 'cookie-session'
            serializer = TaggedJSONSerializer()
            signer_kwargs = {
                'key_derivation': 'hmac',
                'digest_method': hashlib.sha1
            }
            s = URLSafeTimedSerializer(secret_key, salt=salt,
		                              serializer=serializer,
		                              signer_kwargs=signer_kwargs)
            return s.dumps(cookieDict)

print decode_flask_cookie('secret_xxx','.eJwVj0FrhDAUhP9KeWcp0V0vQg9bXEXhJViyhJeLsJqq0bSgXVqz7H9v9jCHmfkOM3doP1ezjZD9rDcTQTv1kN3h5QoZcKUXsqdfLQdGilJtgxwftaudkMWo7XBEe0nI15ZUYck2sbbFLEr6I9UcUBHjwWMesnxcRFkxkqcd1Zk9eV5WO_ohRldP6IuRbD-T_3CYLxP3fBJyTklSKhQGrktQFqGvdvKXAyVNirJLRf4eNs5v8IhgNc64q1nbzXTfX_0G2ZGxVxbBbQvh8xjE8PgHNYdQMA.Diie3A.rteZorgyz6ijPJmNG1oM7XEAfFQ')

print decode_flask_cookie('somexxxkey','.eJxFkE-LgzAUxL_KknMPGvUi9GCJiMJ7IsTKy6WwamuSzS5oi39Kv_u6vex5mPnNzJNdrmM_DSy-j4_-wC66Y_GTfXyymKFMIxSVp1walBJm4rkHXGmVnb9AJjOKkwYHHDkOJGtfZfkMJvHJdBrNbaOtWmG7RZjRXIpqoSbnZOpl17gSbUS8WlUGCwjLwaQ-mJMD0WlyEBA_G9VUPhoIlLAbSlhLaUPk6Z8_ApEEZAqHpvXKpj6y14G103i93H9s__0_QbQhyM6Swb1SYcGlHF29o1OfnHI7YgUJXikLi6JzygwaquM77jH14_sOFvrs9Qujt1-Z.DihxUw.LS2FU2seMhNUhAtdaflbyKKLL8E')

print '----',encode_flask_cookie('somexxxkey',{u'csrf_token': '47817db3b62d2a66e9ca5bffed213492d47ff8b1', u'_fresh': True, u'user_id': u'41', u'_id': '519444fa7930cb43fbdee10040b2c67caa55db0205b7b683c423894f0841ab6b51686d79cd2dc109621520f07bbc7ceced5637d935329987a1689007b2f6749e'})
*/

/*
# login test
from flask import Flask, Response, redirect, url_for, request, session, abort, render_template_string
from flask_login import LoginManager, UserMixin, \
                                login_required, login_user, logout_user
from datetime import timedelta
from flask_wtf.csrf import CSRFProtect
from wtforms.form import Form

csrf = CSRFProtect()
WTF_CSRF_SECRET_KEY = 'asdfadsfad'
app = Flask(__name__)
app.debug=True
csrf.init_app(app)

# config
app.config.update(
    DEBUG = True,
    SECRET_KEY = 'secret_xxx'
)

# flask-login
login_manager = LoginManager()
login_manager.init_app(app)
login_manager.login_view = "login"


# silly user model
class User(UserMixin):

    def __init__(self, id):
        self.id = id
        self.name = "user" + str(id)
        self.password = self.name + "_secret"

    def __repr__(self):
        return "%d/%s/%s" % (self.id, self.name, self.password)


# create some users with ids 1 to 20
users = [User(id) for id in range(1, 21)]


# some protected url
@app.route('/')
@login_required
def home():
    return Response("Hello World!")


# somewhere to login
@app.route("/login", methods=["GET", "POST"])
def login():
    form = Form()
    if request.method == 'POST':
        if form.validate_on_submit():
            username = request.form['username']
            password = request.form['password']
            id = 1
            user = User(id)
            login_user(user,True,timedelta(seconds=400))
            # return redirect(request.args.get("next"))
            return Response('login!')
    else:
        return render_template_string('''
        <form action="" method="post">
            <input type="hidden" name="csrf_token" value="{{ csrf_token() }}" />
            <p><input type=text name=username>
            <p><input type=password name=password>
            <p><input type=submit value=Login>
        </form>
        ''',form=form)


# somewhere to logout
@app.route("/logout")
@login_required
def logout():
    logout_user()
    return Response('<p>Logged out</p>')


# handle login failed
@app.errorhandler(401)
def page_not_found(e):
    return Response('<p>Login failed</p>')


# callback to reload the user object
@login_manager.user_loader
def load_user(userid):
    return User(userid)


# @login_manager.user_loader
# def user_loader(token):
#     print("---user_loader, token=", token)
#     return load_token(token)

if __name__ == "__main__":
    app.run()
*/
