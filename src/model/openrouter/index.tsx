import React from 'react'

const OpenRouterView: React.FC = () => {
  return (
    <div className="model-config-view card-premium animate-fade">
      <div className="config-header">
         <h2>OpenRouter Configuration</h2>
         <p>Link your API key to start using OpenRouter with Dardcor Agent.</p>
      </div>
      
      <div className="config-form">
         <div className="form-group">
            <label>API Key</label>
            <input type="password" placeholder="Enter OpenRouter API Key..." className="input-premium" />
         </div>
         
         <div className="form-group">
            <label>Select Model</label>
            <select className="input-premium" defaultValue="Select a model...">
               <option>Select a model...</option>
               <option>GPT-4o</option>
               <option>Claude 3.5 Sonnet</option>
               <option>Llama 3 70B</option>
            </select>
         </div>

         <button className="btn-primary-glow">Save Configuration</button>
      </div>
    </div>
  )
}

export default OpenRouterView
