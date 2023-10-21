package elevengo

import (
	"errors"
	"fmt"
	"io"

	"github.com/deadblue/elevengo/internal/util"
	"github.com/deadblue/elevengo/lowlevel/api"
	"github.com/deadblue/elevengo/option"
)

const (
	formatQrcodeSession = "%s|%d|%s|%s"
)

// QrcodeSession holds the information during a QRCode login process.
type QrcodeSession struct {
	// QRCode image content.
	Image []byte
	// Hidden fields.
	appType string
	uid     string
	time    int64
	sign    string
}

// Marshal marshals QrcodeSession into a string, which can be tranfered out of
// process.
func (s *QrcodeSession) Marshal() string {
	if s.time == 0 {
		return ""
	}
	return fmt.Sprintf(formatQrcodeSession, s.uid, s.time, s.sign, s.appType)
}

// Unmarshal fills QrcodeSession from a string which is generated by |Marshal|.
func (s *QrcodeSession) Unmarshal(text string) (err error) {
	_, err = fmt.Sscanf(
		text, formatQrcodeSession,
		&s.uid, &s.time, &s.sign, &s.appType,
	)
	return
}

var ErrQrcodeCancelled = errors.New("QRcode cancelled")

// QrcodeStart starts a QRcode sign-in session.
// The session is for web by default, you can change sign-in app by passing a
// "option.QrcodeLoginOption".
//
// Example:
//
//	agent := elevengo.Default()
//	session := elevengo.QrcodeSession()
//	agent.QrcodeStart(session, option.QrcodeLoginLinux)
func (a *Agent) QrcodeStart(session *QrcodeSession, options ...option.QrcodeOption) (err error) {
	// Apply options
	for _, opt := range options {
		switch opt := opt.(type) {
		case option.QrcodeLoginOption:
			session.appType = string(opt)
		}
	}
	if session.appType == "" {
		session.appType = string(option.QrcodeLoginWeb)
	}
	spec := (&api.QrcodeTokenSpec{}).Init(session.appType)
	if err = a.llc.CallApi(spec); err != nil {
		return
	}
	session.uid = spec.Result.Uid
	session.time = spec.Result.Time
	session.sign = spec.Result.Sign
	// Fetch QRcode image data
	var reader io.ReadCloser
	if reader, err = a.Fetch(api.QrcodeImageUrl(session.appType, session.uid)); err != nil {
		return
	}
	defer util.QuietlyClose(reader)
	session.Image, err = io.ReadAll(reader)
	return
}

func (a *Agent) qrcodeSignIn(session *QrcodeSession) (err error) {
	spec := (&api.QrcodeLoginSpec{}).Init(session.appType, session.uid)
	if err = a.llc.CallApi(spec); err != nil {
		return
	}
	return a.afterSignIn(spec.Result.Cookie.UID)
}

// QrcodePoll polls the session state, and automatically sin
func (a *Agent) QrcodePoll(session *QrcodeSession) (done bool, err error) {
	spec := (&api.QrcodeStatusSpec{}).Init(
		session.uid, session.time, session.sign,
	)
	if err = a.llc.CallApi(spec); err != nil {
		return
	}
	switch spec.Result.Status {
	case -2:
		err = ErrQrcodeCancelled
	case 2:
		err = a.qrcodeSignIn(session)
		done = err == nil
	}
	return
}
