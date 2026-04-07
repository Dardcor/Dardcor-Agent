import React, { useState, useEffect, useRef } from 'react'
import wsService from '../services/websocket'

interface CommandResult {
  command: string
  output: string
  exitCode: number
  timestamp: string
}

const Terminal: React.FC = () => {
  const [history, setHistory] = useState<CommandResult[]>([])
  const [input, setInput] = useState('')
  const [isExecuting, setIsExecuting] = useState(false)
  const terminalEndRef = useRef<HTMLDivElement>(null)

  useEffect(() => {
    const unsub = wsService.on('command_result', (msg: any) => {
      setHistory(prev => [...prev, {
        command: msg.payload.command,
        output: msg.payload.output,
        exitCode: msg.payload.exit_code,
        timestamp: new Date().toLocaleTimeString()
      }])
      setIsExecuting(false)
    })

    return () => unsub()
  }, [])

  useEffect(() => {
    terminalEndRef.current?.scrollIntoView({ behavior: 'smooth' })
  }, [history])

  const handleRun = (e: React.FormEvent) => {
    e.preventDefault()
    if (!input.trim() || isExecuting) return

    setIsExecuting(true)
    wsService.send('execute_command', { command: input.trim() })
    setInput('')
  }

  return (
    <div className="terminal-container">
      <div className="terminal-header">
        <div className="terminal-title">
          <span>>_</span> TERMINAL
        </div>
        <div className="terminal-controls">
          <button onClick={() => setHistory([])}>Clear</button>
        </div>
      </div>
      <div className="terminal-output">
        {history.map((item, i) => (
          <div key={i} className="terminal-entry">
            <div className="terminal-cmd-row">
              <span className="prompt">$</span>
              <span className="cmd-text">{item.command}</span>
            </div>
            <pre className={`cmd-output ${item.exitCode !== 0 ? 'error' : ''}`}>
              {item.output}
            </pre>
          </div>
        ))}
        {isExecuting && (
          <div className="terminal-loading">
            <span className="prompt">$</span>
            <span className="cursor-blink">_</span>
          </div>
        )}
        <div ref={terminalEndRef} />
      </div>
      <form className="terminal-input-row" onSubmit={handleRun}>
        <span className="prompt">$</span>
        <input
          value={input}
          onChange={e => setInput(e.target.value)}
          placeholder="Enter command..."
          autoFocus
          spellCheck={false}
          disabled={isExecuting}
        />
      </form>
    </div>
  )
}

export default Terminal
