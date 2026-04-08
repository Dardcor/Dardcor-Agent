import React from 'react'

const GeminiView: React.FC = () => {
  return (
    <div className="model-config-view card-premium animate-fade">
      <div className="config-header">
         <h2>Google Gemini Configuration</h2>
         <p>Link your API key to start using Google Gemini with Dardcor Agent.</p>
      </div>
      
      <div className="config-form">
         <div className="form-group">
            <label>API Key</label>
            <input type="password" placeholder="Enter Gemini API Key..." className="input-premium" />
         </div>
         
         <div className="form-group">
            <label>Select Model</label>
            <select className="input-premium" defaultValue="Select a model...">
               <option>Select a model...</option>
               <option>Gemini 1.5 Pro</option>
               <option>Gemini 1.5 Flash</option>
               <option>Gemini 1.0 Pro</option>
            </select>
         </div>

         <button className="btn-primary-glow">Save Configuration</button>
      </div>
    </div>
  )
}

export default GeminiView
