import React, { useState } from 'react'

const PLANT_TYPES = ['tomato', 'basil', 'succulent']

export default function AddPlant({ onDone }) {
  const [plantId, setPlantId] = useState('')
  const [type, setType] = useState(PLANT_TYPES[0])
  const [error, setError] = useState(null)
  const [status, setStatus] = useState(null)

  async function handleSubmit(e) {
    e.preventDefault()
    setError(null)
    setStatus(null)

    const trimmedId = (plantId || '').trim()
    if (!trimmedId) {
      setError('Plant id is required')
      return
    }

    if (!PLANT_TYPES.includes(type)) {
      setError('Invalid plant type')
      return
    }

    try {
      // Try to POST to the backend if available
      const res = await fetch('/api/add_plant', {
        method: 'POST',
        credentials: 'include',
        headers: { 'Content-Type': 'application/json' },
        body: JSON.stringify({ plantId: trimmedId, type })
      })

      if (res.ok) {
        setStatus('Plant added successfully')
        setPlantId('')
        setType(PLANT_TYPES[0])
        // call onDone to return to previous view if provided
        if (onDone) onDone()
      } else {
        const text = await res.text().catch(() => '')
        setError('Server error: ' + (text || res.statusText || res.status))
      }
    } catch (err) {
      // Network or other error: show success locally but report error
      setError('Network error: ' + err.message)
    }
  }

  return (
    <div className="container">
      <h3>Add a new plant</h3>
      <form onSubmit={handleSubmit}>
        <div style={{ marginBottom: 8 }}>
          <label>
            Plant id (string)
            <br />
            <input
              type="text"
              value={plantId}
              onChange={(e) => setPlantId(e.target.value)}
              placeholder="e.g. plant-123"
            />
          </label>
        </div>

        <div style={{ marginBottom: 8 }}>
          <label>
            Type of plant
            <br />
            <select value={type} onChange={(e) => setType(e.target.value)}>
              {PLANT_TYPES.map(t => (
                <option key={t} value={t}>{t}</option>
              ))}
            </select>
          </label>
        </div>

        <div style={{ marginTop: 12 }}>
          <button type="submit">Add plant</button>
          <button type="button" style={{ marginLeft: 8 }} onClick={() => onDone && onDone()}>Cancel</button>
        </div>

        {error && <p style={{ color: 'crimson' }}>{error}</p>}
        {status && <p style={{ color: 'green' }}>{status}</p>}
      </form>
    </div>
  )
}
