package softphone

import (
	"github.com/google/uuid"
	"github.com/gorilla/websocket"
	"github.com/pion/webrtc/v2"
	"github.com/ringcentral/ringcentral-go"
	"github.com/ringcentral/ringcentral-go/definitions"
	log "github.com/sirupsen/logrus"
	"math/rand"
)

type Softphone struct {
	Device           definitions.SipRegistrationDeviceInfo
	OnTrack          func(track *webrtc.Track)
	OnInvite         func(inviteMessage SipMessage)
	rc               ringcentral.RestClient
	sipInfo          definitions.SIPInfoResponse
	wsConn           *websocket.Conn
	fakeDomain       string
	fakeEmail        string
	fromTag          string
	toTag            string
	callId           string
	cseq             int
	messageListeners map[string]func(string)
	inviteKey        string
}

func NewSoftPhone(rc ringcentral.RestClient) *Softphone {
	configureLog()
	softphone := Softphone{}
	softphone.OnInvite = func(inviteMessage SipMessage) {}
	softphone.OnTrack = func(track *webrtc.Track) {}
	softphone.rc = rc
	softphone.fakeDomain = uuid.New().String() + ".invalid"
	softphone.fakeEmail = uuid.New().String() + "@" + softphone.fakeDomain
	softphone.fromTag = uuid.New().String()
	softphone.toTag = uuid.New().String()
	softphone.callId = uuid.New().String()
	softphone.cseq = rand.Intn(10000) + 1
	softphone.messageListeners = make(map[string]func(string))

	softphone.register()
	return &softphone
}

func (softphone *Softphone) addMessageListener(messageListener func(string)) string {
	key := uuid.New().String()
	softphone.messageListeners[key] = messageListener
	return key
}
func (softphone *Softphone) removeMessageListener(key string) {
	delete(softphone.messageListeners, key)
}

func (softphone *Softphone) request(sipMessage SipMessage, responseHandler func(string) bool) {
	log.Debug("↑↑↑\n", sipMessage.ToString())
	if responseHandler != nil {
		var key string
		key = softphone.addMessageListener(func(message string) {
			done := responseHandler(message)
			if done {
				softphone.removeMessageListener(key)
			}
		})
	}
	err := softphone.wsConn.WriteMessage(1, []byte(sipMessage.ToString()))
	if err != nil {
		log.Fatal(err)
	}
}

func (softphone *Softphone) response(message string) {
	log.Debug("↑↑↑\n", message)
	err := softphone.wsConn.WriteMessage(1, []byte(message))
	if err != nil {
		log.Fatal(err)
	}
}
