package helpers

import (
	"encoding/base64"
	"encoding/json"
	"fmt"

	"github.com/pion/interceptor"
	"github.com/pion/interceptor/pkg/intervalpli"
	"github.com/pion/webrtc/v4"
)

type WebRTC struct {
	SdpChan chan string ``
}

func NewWebRTC() *WebRTC {
	return &WebRTC{
		SdpChan: make(chan string),
	}
}

func (w *WebRTC) Connection() {
	offer := webrtc.SessionDescription{}
	w.decode(<-w.SdpChan, &offer)

	fmt.Println("offer değeri", offer.Type)

	peerConnectionConfig := webrtc.Configuration{
		ICEServers: []webrtc.ICEServer{
			{
				URLs: []string{},
			},
		},
	}

	m := &webrtc.MediaEngine{}

	if err := m.RegisterDefaultCodecs(); err != nil {
		fmt.Println("condeclerde hata oluştu", err)
		return
	}

	i := &interceptor.Registry{}

	if err := webrtc.RegisterDefaultInterceptors(m, i); err != nil {
		panic(err)
	}

	intervalpliFactory, err := intervalpli.NewReceiverInterceptor()
	if err != nil {
		panic(err)
	}
	i.Add(intervalpliFactory)

	peerConnection, _ := webrtc.NewAPI(webrtc.WithMediaEngine(m), webrtc.WithInterceptorRegistry(i)).NewPeerConnection(peerConnectionConfig)
	defer func() {
		err := peerConnection.Close()
		if err != nil {
			fmt.Printf("cannot close peerConnection: %v\n", err)
		}
	}()

	peerConnection.AddTransceiverFromKind(webrtc.RTPCodecTypeVideo)
	localTrackChan := make(chan *webrtc.TrackLocalStaticRTP)

	peerConnection.OnTrack(func(tr *webrtc.TrackRemote, r *webrtc.RTPReceiver) {
		localTrack, _ := webrtc.NewTrackLocalStaticRTP(tr.Codec().RTPCodecCapability, "video", "pion")
		localTrackChan <- localTrack

		rtpBuf := make([]byte, 1400)
		for {
			i, _, _ := tr.Read(rtpBuf)
			localTrack.Write(rtpBuf[:i])
		}
	})

	peerConnection.SetRemoteDescription(offer)

	answer, _ := peerConnection.CreateAnswer(nil)

	gatherComlete := webrtc.GatheringCompletePromise(peerConnection)

	peerConnection.SetLocalDescription(answer)

	<-gatherComlete

	fmt.Println(w.encode(peerConnection.LocalDescription()))

	localTrac := <-localTrackChan
	for {
		recvOnlyOffer := webrtc.SessionDescription{}
		w.decode(<-w.SdpChan, &recvOnlyOffer)
		peerConnection, _ := webrtc.NewPeerConnection(peerConnectionConfig)

		rtpsender, err := peerConnection.AddTrack(localTrac)
		if err != nil {
			panic(err)
		}

		go func() {
			rtcpBuf := make([]byte, 1500)
			for {
				rtpsender.Read(rtcpBuf)
			}
		}()

		peerConnection.SetRemoteDescription(recvOnlyOffer)
		answer, _ := peerConnection.CreateAnswer(nil)

		gatherComlete = webrtc.GatheringCompletePromise(peerConnection)

		peerConnection.SetLocalDescription(answer)

		<-gatherComlete
		fmt.Println(w.encode(peerConnection.LocalDescription()))

	}
}

func (w *WebRTC) decode(in string, obj *webrtc.SessionDescription) {
	b, err := base64.StdEncoding.DecodeString(in)
	if err != nil {
		panic(err)
	}

	if err = json.Unmarshal(b, obj); err != nil {
		panic(err)
	}
}

func (w *WebRTC) encode(obj *webrtc.SessionDescription) string {
	b, _ := json.Marshal(&obj)
	return base64.StdEncoding.EncodeToString(b)
}
