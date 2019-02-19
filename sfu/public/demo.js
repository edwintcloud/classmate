/* eslint-env browser */
var log = msg => {
  document.getElementById('logs').innerHTML += msg + '<br>'
}

window.createSession = isPublisher => {
  let pc = new RTCPeerConnection({
    iceServers: [
      {
        urls: 'stun:stun.l.google.com:19302'
      }
    ]
  })
  pc.oniceconnectionstatechange = e => log(pc.iceConnectionState)
  pc.onicecandidate = event => {
    if (event.candidate === null) {
      fetch("http://localhost:8080/sdp", {
        method: 'post',
        headers: {
          'Content-Type': 'application/json'
        },
        body: btoa(JSON.stringify(pc.localDescription))
      }).then(res => {
        return res.json();
      }).then(data => {
        startSession(data.answer);
      });
    }
  }

  if (isPublisher) {
    navigator.mediaDevices.getDisplayMedia({video: true, audio: false})
      .then(stream => pc.addStream(document.getElementById('video1').srcObject = stream))
      .catch(log)
    pc.onnegotiationneeded = e => {
      pc.createOffer()
        .then(d => pc.setLocalDescription(d))
        .catch(log)
    }
    document.getElementById('signalingContainer').style = 'display: block'
  } else {
    pc.createOffer({offerToReceiveVideo: true})
      .then(d => pc.setLocalDescription(d))
      .catch(log)

    pc.ontrack = function (event) {
      var el = document.getElementById('video1')
      el.srcObject = event.streams[0]
      el.autoplay = true
      el.controls = true
    }
  }

    const startSession = (answer) => {
    let sd = answer
    if (sd === '') {
      return alert('Session Description must not be empty')
    }

    try {
      pc.setRemoteDescription(new RTCSessionDescription(JSON.parse(atob(sd))))
    } catch (e) {
      alert(e)
    }
  }

  let btns = document.getElementsByClassName('createSessionButton')
  for (let i = 0; i < btns.length; i++) {
    btns[i].style = 'display: none'
  }
}

const testClient = () => {
  client = window.open("http://localhost:8080/#/client", '_blank').focus()
}

function isClient() {
  if (window.location.hash) {
    window.createSession(false);
  }
}