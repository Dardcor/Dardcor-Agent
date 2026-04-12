import React, { useState, useEffect, useRef } from 'react'
import wsService from '../services/websocket'

interface TerminalLog {
  type: 'command' | 'output' | 'error' | 'info'
  text: string
  shell?: string
}

const Terminal: React.FC = () => {
  const [logs, setLogs] = useState<TerminalLog[]>([])
  const [input, setInput] = useState('')
  const [isExecuting, setIsExecuting] = useState(false)
  const [activeCommandId, setActiveCommandId] = useState<string | null>(null)
  const [activeShell, setActiveShell] = useState('cmd')
  const terminalEndRef = useRef<HTMLDivElement>(null)

  useEffect(() => {
    const isWindows = navigator.platform.toLowerCase().includes('win')
    setActiveShell(isWindows ? 'powershell' : 'bash')

    const unsubOutput = wsService.on('command_output', (msg: any) => {
      if (msg.payload.command_id) {
         setActiveCommandId(msg.payload.command_id)
      }
      setLogs(prev => [...prev, {
        type: msg.payload.is_error ? 'error' : 'output',
        text: msg.payload.output
      }])
    })

    const unsubComplete = wsService.on('command_complete', (msg: any) => {
      setIsExecuting(false)
      setActiveCommandId(null)
      setLogs(prev => [...prev, {
        type: 'info',
        text: `\n[Process exited with code ${msg.payload.exit_code}]`
      }])
    })

    const unsubError = wsService.on('error', (msg: any) => {
       if (isExecuting) {
          setIsExecuting(false)
          setLogs(prev => [...prev, { type: 'error', text: `System Error: ${msg.payload.error}` }])
       }
    })

    return () => {
       unsubOutput()
       unsubComplete()
       unsubError()
    }
  }, [isExecuting])

  useEffect(() => {
    terminalEndRef.current?.scrollIntoView({ behavior: 'smooth' })
  }, [logs])

  const handleRun = (e: React.FormEvent) => {
    e.preventDefault()
    if (!input.trim() || isExecuting) return

    const cmd = input.trim()
    setLogs(prev => [...prev, { type: 'command', text: cmd, shell: activeShell }])
    setIsExecuting(true)
    setActiveCommandId(null)

    wsService.send('execute_command', {
       command: cmd,
       shell: activeShell
    })
    setInput('')
  }

  const handleStop = () => {
     if (!isExecuting || !activeCommandId) return
     wsService.send('kill_command', { id: activeCommandId })
     setIsExecuting(false)
     setActiveCommandId(null)
     setLogs(prev => [...prev, { type: 'info', text: '\n[Process terminated by user]' }])
  }

  const clearLogs = () => setLogs([])

  return (
    <div className="terminal-container">
      <div className="terminal-header">
        <div className="terminal-title">
          <span>DARDCOR TERMINAL</span>
        </div>
        <div className="terminal-actions">
           <select
             className="shell-selector"
             value={activeShell}
             onChange={(e) => setActiveShell(e.target.value)}
             disabled={isExecuting}
           >
              <option value="powershell">PowerShell</option>
              <option value="cmd">CMD</option>
              <option value="bash">Bash</option>
              <option value="sh">Sh</option>
              <option value="zsh">Zsh</option>
           </select>
           {isExecuting && (
              <button className="terminal-action-btn stop" onClick={handleStop}>Stop</button>
           )}
           <button className="terminal-action-btn" onClick={clearLogs}>Clear</button>
        </div>
      </div>

      <div className="terminal-body" onClick={() => document.getElementById('term-input')?.focus()}>
        <div className="terminal-history">
           {logs.map((log, i) => (
              <div key={i} className={`terminal-line ${log.type}`}>
                 {log.type === 'command' && (
                    <span className="terminal-prompt">
                       <span className="prompt-shell">[{log.shell}]</span>
                       <span className="prompt-symbol">▶</span>
                    </span>
                 )}
                 <span className="terminal-text">{log.text}</span>
              </div>
           ))}
           {isExecuting && (
              <div className="terminal-line info">
                 <span className="terminal-spinner">⠋</span>
                 <span className="terminal-text">Executing...</span>
              </div>
           )}
           <div ref={terminalEndRef} />
        </div>

        <form className="terminal-input-row" onSubmit={handleRun}>
          <span className="terminal-prompt">
             <span className="prompt-shell">[{activeShell}]</span>
             <span className="prompt-symbol">▶</span>
          </span>
          <input
            id="term-input"
            className="terminal-input"
            value={input}
            onChange={e => setInput(e.target.value)}
            disabled={isExecuting}
            autoFocus
            autoComplete="off"
            spellCheck={false}
          />
        </form>
      </div>
    </div>
  )
}

export default Terminal




