package main

import (
	"encoding/base64"
	"encoding/json"
	"fmt"
	"io/ioutil"
	"log"
	"net/http"
	"strings"
	"sync"
	"time"

	"github.com/pions/rtcp"
	"github.com/pions/rtp"
	"github.com/pions/webrtc"
	"github.com/pions/webrtc/pkg/ice"
)

// Response is the model for our json http responses
type Response struct {
	Message string `json:"message"`
	Status  int    `json:"status"`
}

// make channels we will need
var peer = make(chan string)
var initiator = make(chan string)
var peerAns = make(chan string)
var initiatorAns = make(chan string)
var inboundSSRC = make(chan uint32)
var inboundPayloadType = make(chan uint8)
var outboundRTP = []chan<- *rtp.Packet{}
var outboundRTPLock sync.RWMutex

// ice: Interactive Connection Establishment RFC 5245
// each client sends connectivity ping every 20ms
// offerer is the controlling agent (the go server)
// uses udp protocol
// STUN : session traversal utilities for network address translation RFC 5389
var connectionConfig = webrtc.Configuration{
	ICEServers: []webrtc.ICEServer{
		{
			URLs: []string{"stun:stun.l.google.com:19302"},
		},
	},
}

func main() {

	// handle / route by serving files in public
	http.Handle("/", http.FileServer(http.Dir("public")))

	// handle sdp peer requests
	http.HandleFunc("/peer", func(w http.ResponseWriter, r *http.Request) {
		fmt.Println("peer")

		// read req body buffer into byte slice
		body, err := ioutil.ReadAll(r.Body)
		if err != nil {
			w.WriteHeader(http.StatusBadRequest)
			json.NewEncoder(w).Encode(Response{
				Message: err.Error(),
				Status:  http.StatusBadRequest,
			})
		}

		// send string body to peer channel
		peer <- string(body)

		// wait for answer from offerer
		answer := <-peerAns

		// if answer contains error, respond with error
		if strings.Contains(answer, "error:") {
			json.NewEncoder(w).Encode(Response{
				Message: strings.Replace(answer, "error: ", "", 1),
				Status:  http.StatusBadRequest,
			})
		} else {
			// respond with answer
			json.NewEncoder(w).Encode(Response{
				Message: answer,
				Status:  http.StatusOK,
			})
		}
	})

	// handle sdp initiator request
	http.HandleFunc("/initiator", func(w http.ResponseWriter, r *http.Request) {

		// read req body buffer into byte slice
		body, err := ioutil.ReadAll(r.Body)
		if err != nil {
			w.WriteHeader(http.StatusBadRequest)
			json.NewEncoder(w).Encode(Response{
				Message: err.Error(),
				Status:  http.StatusBadRequest,
			})
		}

		// send string body to peer channel
		initiator <- string(body)

		// wait for answer from offerer
		answer := <-initiatorAns

		// if answer contains error, respond with error
		if strings.Contains(answer, "error:") {
			json.NewEncoder(w).Encode(Response{
				Message: strings.Replace(answer, "error: ", "", 1),
				Status:  http.StatusBadRequest,
			})
		} else {
			// respond with answer
			json.NewEncoder(w).Encode(Response{
				Message: answer,
				Status:  http.StatusOK,
			})
		}
	})

	// wait for initiator connections in a new routine
	go waitForInitiators()

	// start http server and listen for requests
	fmt.Println("Starting http server on port", 8080)
	err := http.ListenAndServe(":8080", nil)
	if err != nil {
		log.Fatalf("Unable to start http server: " + err.Error())
	}
}

func waitForInitiators() {
	for {

		offer := webrtc.SessionDescription{}

		// wait from a message to be sent to the sdp channel
		message := <-initiator

		// decode message
		decodedMessage, err := base64.StdEncoding.DecodeString(message)
		if err != nil {
			initiatorAns <- "error: " + err.Error()
			continue
		}

		// unmarshal decodedMessage into offer struct
		err = json.Unmarshal(decodedMessage, &offer)
		if err != nil {
			initiatorAns <- "error: " + err.Error()
			continue
		}

		// setup transport to use VP8 codec
		webrtc.RegisterCodec(webrtc.NewRTPVP8Codec(webrtc.DefaultPayloadTypeVP8, 90000))

		// create new rtc peer connection
		initiatorConnection, err := webrtc.NewPeerConnection(connectionConfig)
		if err != nil {
			initiatorAns <- "error: " + err.Error()
			continue
		}

		// setup handler for when new initiator track starts
		// handler will distribute packets to connected peers
		initiatorConnection.OnTrack(func(track *webrtc.Track) {

			// Send a PLI on an interval so that the publisher is pushing a keyframe every rtcpPLIInterval
			// This can be less wasteful by processing incoming RTCP events, then we would emit a NACK/PLI when a viewer requests it
			go func() {
				ticker := time.NewTicker(time.Second * 3)
				for range ticker.C {
					if err := initiatorConnection.SendRTCP(&rtcp.PictureLossIndication{MediaSSRC: track.SSRC}); err != nil {
						fmt.Println(err)
						return
					}
				}
			}()

			// send track SSRC and PayloadType onto respective channels
			inboundSSRC <- track.SSRC
			inboundPayloadType <- track.PayloadType

			// continuously send packets to connected clients
			// until stopPackets is true
			for {
				rtpPacket := <-track.Packets

				outboundRTPLock.RLock()
				for _, outChan := range outboundRTP {
					outPacket := rtpPacket
					outPacket.Payload = append([]byte{}, outPacket.Payload...)
					select {
					case outChan <- outPacket:
					default:
					}
				}
				outboundRTPLock.RUnlock()
			}
		})

		// state change listener
		initiatorConnection.OnICEConnectionStateChange(func(s ice.ConnectionState) {
			if s == ice.ConnectionStateDisconnected {
				initiatorConnection.Close()
			}
		})

		// set initiator SessionDescription
		err = initiatorConnection.SetRemoteDescription(offer)
		if err != nil {
			initiatorAns <- "error: " + err.Error()
			continue
		}

		// create answer
		answer, err := initiatorConnection.CreateAnswer(nil)
		if err != nil {
			initiatorAns <- "error: " + err.Error()
			continue
		}

		// set local description and start udp listeners
		err = initiatorConnection.SetLocalDescription(answer)
		if err != nil {
			initiatorAns <- "error: " + err.Error()
			continue
		}

		// convert json answer into byte slice
		bytes, err := json.Marshal(answer)
		if err != nil {
			initiatorAns <- "error: " + err.Error()
			continue
		}

		// encode answer to base64 and send to ans channel
		initiatorAns <- base64.StdEncoding.EncodeToString(bytes)

		waitForPeers()
	}
}

func waitForPeers() {

	// wait from inbound ssrc and payload from initiator track
	outboundSSRC := <-inboundSSRC
	outboundPayloadType := <-inboundPayloadType

	for {

		offer := webrtc.SessionDescription{}

		// wait from a message to be sent to the sdp channel
		message := <-peer

		// decode message
		decodedMessage, err := base64.StdEncoding.DecodeString(message)
		if err != nil {
			peerAns <- "error: " + err.Error()
			continue
		}

		// unmarshal decodedMessage into offer struct
		err = json.Unmarshal(decodedMessage, &offer)
		if err != nil {
			peerAns <- "error: " + err.Error()
			continue
		}

		// create new peer connection
		peerConnection, err := webrtc.NewPeerConnection(connectionConfig)
		if err != nil {
			peerAns <- "error: " + err.Error()
			continue
		}

		// Create a single VP8 Track to send video
		vp8Track, err := peerConnection.NewRawRTPTrack(outboundPayloadType, outboundSSRC, "video", "pion")
		if err != nil {
			peerAns <- "error: " + err.Error()
			continue
		}

		// Add track to peer connection
		_, err = peerConnection.AddTrack(vp8Track)
		if err != nil {
			peerAns <- "error: " + err.Error()
			continue
		}

		outboundRTPLock.Lock()
		outboundRTP = append(outboundRTP, vp8Track.RawRTP)
		outboundRTPLock.Unlock()

		// set initiator SessionDescription
		err = peerConnection.SetRemoteDescription(offer)
		if err != nil {
			initiatorAns <- "error: " + err.Error()
			continue
		}

		// create answer
		answer, err := peerConnection.CreateAnswer(nil)
		if err != nil {
			initiatorAns <- "error: " + err.Error()
			continue
		}

		// convert json answer into byte slice
		bytes, err := json.Marshal(answer)
		if err != nil {
			initiatorAns <- "error: " + err.Error()
			continue
		}

		// encode answer to base64 and send to ans channel
		peerAns <- base64.StdEncoding.EncodeToString(bytes)

	}
}
