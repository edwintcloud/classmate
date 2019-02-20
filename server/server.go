package main

import (
	"encoding/base64"
	"encoding/json"
	"errors"
	"net/http"
	"sync"
	"time"

	"github.com/labstack/echo"
	"github.com/labstack/echo/middleware"
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
var hostSSRC uint32
var hostPayloadType uint8
var outboundRTP = []chan<- *rtp.Packet{}
var mutex sync.RWMutex

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

	// register echo instance
	e := echo.New()

	// register logging middleware with echo instance
	e.Use(middleware.LoggerWithConfig(middleware.LoggerConfig{
		Format: "${method}  ${uri}  ${latency_human}  ${status}\n",
	}))

	// serve public folder as /
	e.Static("/", "public")

	// host connection
	e.POST("/host", func(c echo.Context) error {
		body := Response{}
		err := c.Bind(&body)
		if err != nil {
			return c.JSON(http.StatusBadRequest, Response{
				Message: err.Error(),
				Status:  http.StatusBadRequest,
			})
		}
		answer, err := startHost(body.Message)
		if err != nil {
			return c.JSON(http.StatusBadRequest, Response{
				Message: err.Error(),
				Status:  http.StatusBadRequest,
			})
		}

		return c.JSON(http.StatusOK, Response{
			Message: answer,
			Status:  http.StatusOK,
		})
	})

	// peer connection
	e.POST("/peer", func(c echo.Context) error {
		body := Response{}
		err := c.Bind(&body)
		if err != nil {
			return c.JSON(http.StatusBadRequest, Response{
				Message: err.Error(),
				Status:  http.StatusBadRequest,
			})
		}
		answer, err := connectPeer(body.Message)
		if err != nil {
			return c.JSON(http.StatusBadRequest, Response{
				Message: err.Error(),
				Status:  http.StatusBadRequest,
			})
		}

		return c.JSON(http.StatusOK, Response{
			Message: answer,
			Status:  http.StatusOK,
		})
	})

	// Start server
	e.Logger.Fatal(e.Start(":8080"))

}

func startHost(clientKey string) (string, error) {
	offer := webrtc.SessionDescription{}

	// decode message
	decodedMessage, err := base64.StdEncoding.DecodeString(clientKey)
	if err != nil {
		return "", err
	}

	// unmarshal decodedMessage into offer struct
	err = json.Unmarshal(decodedMessage, &offer)
	if err != nil {
		return "", err
	}

	// setup transport to use VP8 codec
	webrtc.RegisterCodec(webrtc.NewRTPVP8Codec(webrtc.DefaultPayloadTypeVP8, 90000))

	// create new rtc peer connection
	initiatorConnection, err := webrtc.NewPeerConnection(connectionConfig)
	if err != nil {
		return "", err
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
					return
				}
			}
		}()

		// send track SSRC and PayloadType onto respective channels
		mutex.Lock()
		hostSSRC = track.SSRC
		hostPayloadType = track.PayloadType
		mutex.Unlock()
		// continuously send packets to connected clients
		// until stopPackets is true
		for {
			rtpPacket := <-track.Packets

			mutex.RLock()
			for _, outChan := range outboundRTP {
				outPacket := rtpPacket
				outPacket.Payload = append([]byte{}, outPacket.Payload...)
				select {
				case outChan <- outPacket:
				default:
				}
			}
			mutex.RUnlock()
		}
	})

	// state change listener
	initiatorConnection.OnICEConnectionStateChange(func(s ice.ConnectionState) {
		if s == ice.ConnectionStateDisconnected {
			return
		}
	})

	// set initiator SessionDescription
	err = initiatorConnection.SetRemoteDescription(offer)
	if err != nil {
		return "", err
	}

	// create answer
	answer, err := initiatorConnection.CreateAnswer(nil)
	if err != nil {
		return "", err
	}

	// set local description and start udp listeners
	err = initiatorConnection.SetLocalDescription(answer)
	if err != nil {
		return "", err
	}

	// convert json answer into byte slice
	bytes, err := json.Marshal(answer)
	if err != nil {
		return "", err
	}

	// encode answer to base64 and send to ans channel
	return base64.StdEncoding.EncodeToString(bytes), nil
}

func connectPeer(clientKey string) (string, error) {

	// if no host available then return error
	if hostPayloadType == 0 || hostSSRC == 0 {
		return "", errors.New("no hosts available")
	}

	// connect to first host in channel
	offer := webrtc.SessionDescription{}

	// decode message
	decodedMessage, err := base64.StdEncoding.DecodeString(clientKey)
	if err != nil {
		return "", err
	}

	// unmarshal decodedMessage into offer struct
	err = json.Unmarshal(decodedMessage, &offer)
	if err != nil {
		return "", err
	}

	// create new peer connection
	peerConnection, err := webrtc.NewPeerConnection(connectionConfig)
	if err != nil {
		return "", err
	}

	// Create a single VP8 Track to send video
	vp8Track, err := peerConnection.NewRawRTPTrack(hostPayloadType, hostSSRC, "video", "pion")
	if err != nil {
		return "", err
	}

	// Add track to peer connection
	_, err = peerConnection.AddTrack(vp8Track)
	if err != nil {
		return "", err
	}

	// append track to outboundrtp channel so host can stream to new peer
	mutex.Lock()
	outboundRTP = append(outboundRTP, vp8Track.RawRTP)
	mutex.Unlock()

	// set initiator SessionDescription
	err = peerConnection.SetRemoteDescription(offer)
	if err != nil {
		return "", err
	}

	// create answer
	answer, err := peerConnection.CreateAnswer(nil)
	if err != nil {
		return "", err
	}

	// convert json answer into byte slice
	bytes, err := json.Marshal(answer)
	if err != nil {
		return "", err
	}

	// encode answer to base64 and send to ans channel
	return base64.StdEncoding.EncodeToString(bytes), nil
}
