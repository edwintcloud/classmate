import React, { Component } from "react";
import Axios from "axios";

class App extends Component {
  constructor(props) {
    super(props);

    this.ws = null;
    this.ip = null;

    this.initialState = {
      message: "",
      messages: []
    };

    this.state = this.initialState;
  }

  componentWillMount() {
    this.ws = new WebSocket(`ws://localhost:8000/ws`);
    this.ws.addEventListener("message", e => {
      const data = JSON.parse(e.data);

      if (data.hasOwnProperty("ip")) {
        console.log(data);
        this.setState({
          messages: [
            ...this.state.messages,
            {
              IP: data.ip,
              Message: data.message
            }
          ]
        });
      } else {
        for (var ip in data) {
          data[ip].forEach(message => {
            this.setState({
              messages: [
                ...this.state.messages,
                {
                  IP: ip,
                  Message: message
                }
              ]
            });
          })
        }
      }
    });

    // get client ip
    Axios.get("http://ip-api.com/json").then(res => {
      this.ip = res.data.query;
    });
  }

  render() {
    const sendMessage = e => {
      e.preventDefault();
      console.log(this.ip);
      this.ws.send(
        JSON.stringify({
          ip: this.ip,
          message: this.state.message
        })
      );
      this.setState({
        message: ""
      });
    };

    const messageChanged = e => {
      this.setState({
        message: e.target.value
      });
    };

    return (
      <div style={styles.App}>
        <form onSubmit={sendMessage}>
          <label for="message">Enter Message: </label>
          <textarea id="message" onChange={messageChanged} />
          <button type="submit">Send</button>
        </form>
        <div style={styles.Messages}>
          {(this.state.messages.length > 0 &&
            this.state.messages.map((message, index) => (
              <p key={index}>
                {message.IP} - {message.Message}
              </p>
            ))) || <p>No Messages</p>}
        </div>
      </div>
    );
  }
}

const styles = {
  App: {
    padding: "20px",
    fontSize: "3em"
  },
  Messages: {
    padding: "20px"
  }
};

export default App;
