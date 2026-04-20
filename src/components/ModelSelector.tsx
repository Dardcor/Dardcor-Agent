import React, { useState, useEffect } from 'react'
import { useParams, useNavigate } from 'react-router-dom'
import AntigravityView from '../model/antigravity'
import GeminiView from '../model/gemini'
import OpenRouterView from '../model/openrouter'
import NvidiaView from '../model/nvidia'
import AnthropicView from '../model/anthropic'
import OpenAIView from '../model/openai'
import DeepSeekView from '../model/deepseek'

type ProviderType = 'antigravity' | 'gemini' | 'openrouter' | 'nvidia' | 'anthropic' | 'openai' | 'deepseek';

const ModelSelector: React.FC = () => {
  const { provider } = useParams<{ provider: string }>()
  const navigate = useNavigate()
  const [activeTab, setActiveTab] = useState<ProviderType>('antigravity')

  useEffect(() => {
    const validProviders: ProviderType[] = ['antigravity', 'gemini', 'openrouter', 'nvidia', 'anthropic', 'openai', 'deepseek'];
    if (provider && validProviders.includes(provider as ProviderType)) {
      setActiveTab(provider as ProviderType)
    }
  }, [provider])

  const handleTabChange = (tab: ProviderType) => {
    setActiveTab(tab)
    navigate(`/model/${tab}`)
  }

  const renderContent = () => {
    switch(activeTab) {
      case 'antigravity': return <AntigravityView />
      case 'gemini':      return <GeminiView />
      case 'openrouter':  return <OpenRouterView />
      case 'nvidia':      return <NvidiaView />
      case 'anthropic':   return <AnthropicView />
      case 'openai':      return <OpenAIView />
      case 'deepseek':    return <DeepSeekView />
    }
  }

  return (
    <div className="model-selector-page">
      <div className="model-navbar">
        <button className={activeTab === 'antigravity' ? 'active' : ''} onClick={() => handleTabChange('antigravity')}>Antigravity</button>
        <button className={activeTab === 'gemini' ? 'active' : ''} onClick={() => handleTabChange('gemini')}>Gemini</button>
        <button className={activeTab === 'openrouter' ? 'active' : ''} onClick={() => handleTabChange('openrouter')}>OpenRouter</button>
        <button className={activeTab === 'nvidia' ? 'active' : ''} onClick={() => handleTabChange('nvidia')}>NVIDIA</button>
        <button className={activeTab === 'anthropic' ? 'active' : ''} onClick={() => handleTabChange('anthropic')}>Anthropic</button>
        <button className={activeTab === 'openai' ? 'active' : ''} onClick={() => handleTabChange('openai')}>OpenAI</button>
        <button className={activeTab === 'deepseek' ? 'active' : ''} onClick={() => handleTabChange('deepseek')}>DeepSeek</button>
      </div>

      <div className="model-content">
        {renderContent()}
      </div>
    </div>
  )
}

export default ModelSelector




