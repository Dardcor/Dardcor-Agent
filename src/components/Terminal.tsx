import React, { useState, useRef, useEffect } from 'react'
import { useTerminal } from '../hooks/useAgent'

const Terminal: React.FC = () => {
  const { output, isRunning, workingDir, setWorkingDir, executeCommand, clearOutput, cancelCommand } = useTerminal()
  const [input, setInput] = useState('')
  const [cmdHistory, setCmdHistory] = useState<string[]>([])
  const [historyIndex, setHistoryIndex] = useState(-1)
  const outputEndRef = useRef<HTMLDivElement>(null)
  const inputRef = useRef<HTMLInputElement>(null)

  useEffect(() => {
    outputEndRef.current?.scrollIntoView({ behavior: 'smooth' })
  }, [output])

  useEffect(() => {
    inputRef.current?.focus()
  }, [])

  const handleSubmit = (e: React.FormEvent) => {
    e.preventDefault()
    if (!input.trim() || isRunning) return

    const cmd = input.trim()

    // Handle built-in commands
    if (cmd === 'clear' || cmd === 'cls') {
      clearOutput()
      setInput('')
      return
    }

    // Handle cd command
    if (cmd.startsWith('cd ')) {
      const newDir = cmd.substring(3).trim()
      setWorkingDir(newDir)
      setInput('')
      return
    }

    setCmdHistory(prev => [...prev, cmd])
    setHistoryIndex(-1)
    executeCommand(cmd)
    setInput('')
  }

  const handleKeyDown = (e: React.KeyboardEvent<HTMLInputElement>) => {
    if (e.ctrlKey && e.key.toLowerCase() === 'c') {
      e.preventDefault()
      if (isRunning && cancelCommand) {
        cancelCommand()
      }
      return
    }

    if (e.key === 'ArrowUp') {
      e.preventDefault()
      if (cmdHistory.length > 0) {
        const newIndex = historyIndex < cmdHistory.length - 1 ? historyIndex + 1 : historyIndex
        setHistoryIndex(newIndex)
        setInput(cmdHistory[cmdHistory.length - 1 - newIndex] || '')
      }
    } else if (e.key === 'ArrowDown') {
      e.preventDefault()
      if (historyIndex > 0) {
        const newIndex = historyIndex - 1
        setHistoryIndex(newIndex)
        setInput(cmdHistory[cmdHistory.length - 1 - newIndex] || '')
      } else {
        setHistoryIndex(-1)
        setInput('')
      }
    }
  }

  const renderLine = (line: string, index: number) => {
    let className = 'terminal-line'
    if (line.startsWith('$ ')) className += ' command'
    else if (line.startsWith('Error:') || line.startsWith('error:')) className += ' error'

    return (
      <div key={index} className={className}>
        {line}
      </div>
    )
  }

  return (
    <div className="terminal-container" onClick={() => inputRef.current?.focus()}>
      {/* Terminal Header */}
      <div className="terminal-toolbar">
        <div className="terminal-toolbar-left">
          <span className="terminal-toolbar-title">Terminal — {workingDir}</span>
        </div>
        <div className="terminal-toolbar-right">
          <button className="terminal-toolbar-btn" onClick={clearOutput} title="Clear" id="terminal-clear">
            Clear
          </button>
        </div>
      </div>

      {/* Output */}
      <div className="terminal-output">
        {output.length === 0 ? (
          <div className="terminal-line info">
            Dardcor Agent Terminal v2.0{'\n'}
            Ketik perintah dan tekan Enter. Ketik "clear" untuk membersihkan.{'\n'}
            Working Directory: {workingDir}{'\n'}
          </div>
        ) : (
          output.map((line, i) => renderLine(line, i))
        )}
        {isRunning && (
          <div className="terminal-line info">
            <span className="terminal-spinner">⣿</span> Running...
          </div>
        )}
        <div ref={outputEndRef} />
      </div>

      {/* Input */}
      <form onSubmit={handleSubmit} className="terminal-input-container">
        <span className="terminal-prompt">❯</span>
        <input
          ref={inputRef}
          type="text"
          className="terminal-input"
          value={input}
          onChange={(e) => setInput(e.target.value)}
          onKeyDown={handleKeyDown}
          placeholder="Ketik perintah..."
          disabled={isRunning}
          autoComplete="off"
          spellCheck={false}
          id="terminal-input"
        />
      </form>
    </div>
  )
}

export default Terminal
