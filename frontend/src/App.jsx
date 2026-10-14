import { useEffect, useRef, useState } from 'react'
import './App.css'

// Each tab picks one of these colors at random
const COLORS = ['#ffd6d6', '#d6f5d6', '#d6e4ff', '#fff3c4', '#f0d6ff', '#ffe0c2']
const myColor = COLORS[Math.floor(Math.random() * COLORS.length)]

function App() {
  const [messages, setMessages] = useState([]) // all chat messages
  const [input, setInput] = useState('')       // text in the input box
  const [name, setName] = useState('')         // user's name
  const [connected, setConnected] = useState(false)
  const [onlineCount, setOnlineCount] = useState(0) // how many users are online
  const socketRef = useRef(null)               // holds the WebSocket object

  useEffect(() => {
    // 1. Open a connection to the Go server
    const socket = new WebSocket('ws://localhost:8080/ws')
    socketRef.current = socket

    // 2. Connection opened
    socket.onopen = () => setConnected(true)

    // 3. A message arrived from the server
    socket.onmessage = (event) => {
      const data = JSON.parse(event.data) // turn text back into an object

      // The server sends two kinds of messages: user count, or a chat message
      if (data.type === 'count') {
        setOnlineCount(data.count)
      } else {
        setMessages((prev) => [...prev, data])
      }
    }

    // 4. Connection closed
    socket.onclose = () => setConnected(false)

    // 5. Close the connection when the component is removed
    return () => socket.close()
  }, [])

  const sendMessage = () => {
    if (input.trim() === '') return
    // Send the name, text and our color together as JSON
    const message = { name: name || 'Anonymous', text: input, color: myColor }
    socketRef.current.send(JSON.stringify(message))
    setInput('')
  }

  return (
    <div className="chat">
      <h2>Go + React WebSocket Chat</h2>
      <p>Status: {connected ? '🟢 Connected' : '🔴 Disconnected'}</p>
      <p>Online users: {onlineCount}</p>
      <p>
        Your color: <span className="color-box" style={{ background: myColor }}></span>
      </p>

      <input
        className="name-input"
        value={name}
        onChange={(e) => setName(e.target.value)}
        placeholder="Enter your name..."
      />

      <div className="messages">
        {messages.map((msg, index) => (
          <div key={index} className="message" style={{ background: msg.color }}>
            <strong>{msg.name}:</strong> {msg.text}
          </div>
        ))}
      </div>

      <div className="input-row">
        <input
          value={input}
          onChange={(e) => setInput(e.target.value)}
          onKeyDown={(e) => e.key === 'Enter' && sendMessage()}
          placeholder="Type a message..."
        />
        <button onClick={sendMessage} disabled={!connected}>Send</button>
      </div>
    </div>
  )
}

export default App
