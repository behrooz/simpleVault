import { useState, useEffect } from 'react'
import './App.css'

const API_BASE_URL = import.meta.env.VITE_API_URL || 'http://localhost:8080/api/v1'

function App() {
  const [secrets, setSecrets] = useState([])
  const [loading, setLoading] = useState(true)
  const [error, setError] = useState(null)
  const [showCreateModal, setShowCreateModal] = useState(false)
  const [showEditModal, setShowEditModal] = useState(false)
  const [selectedSecret, setSelectedSecret] = useState(null)
  const [formData, setFormData] = useState({
    name: '',
    description: '',
    data: {}
  })
  const [newKeyValue, setNewKeyValue] = useState({ key: '', value: '' })

  useEffect(() => {
    fetchSecrets()
  }, [])

  const fetchSecrets = async () => {
    try {
      setLoading(true)
      const response = await fetch(`${API_BASE_URL}/secrets`)
      if (!response.ok) throw new Error('Failed to fetch secrets')
      const result = await response.json()
      setSecrets(result.secrets || [])
      setError(null)
    } catch (err) {
      setError(err.message)
    } finally {
      setLoading(false)
    }
  }

  const handleCreate = async (e) => {
    e.preventDefault()
    try {
      const response = await fetch(`${API_BASE_URL}/secrets`, {
        method: 'POST',
        headers: { 'Content-Type': 'application/json' },
        body: JSON.stringify(formData)
      })
      if (!response.ok) throw new Error('Failed to create secret')
      await fetchSecrets()
      setShowCreateModal(false)
      setFormData({ name: '', description: '', data: {} })
      setError(null)
    } catch (err) {
      setError(err.message)
    }
  }

  const handleUpdate = async (e) => {
    e.preventDefault()
    try {
      const response = await fetch(`${API_BASE_URL}/secrets/${selectedSecret.id}`, {
        method: 'PUT',
        headers: { 'Content-Type': 'application/json' },
        body: JSON.stringify(formData)
      })
      if (!response.ok) throw new Error('Failed to update secret')
      await fetchSecrets()
      setShowEditModal(false)
      setSelectedSecret(null)
      setFormData({ name: '', description: '', data: {} })
      setError(null)
    } catch (err) {
      setError(err.message)
    }
  }

  const handleDelete = async (id) => {
    if (!confirm('Are you sure you want to delete this secret?')) return
    try {
      const response = await fetch(`${API_BASE_URL}/secrets/${id}`, {
        method: 'DELETE'
      })
      if (!response.ok) throw new Error('Failed to delete secret')
      await fetchSecrets()
      setError(null)
    } catch (err) {
      setError(err.message)
    }
  }

  const openEditModal = (secret) => {
    setSelectedSecret(secret)
    setFormData({
      name: secret.name,
      description: secret.description,
      data: { ...secret.data }
    })
    setShowEditModal(true)
  }

  const addKeyValue = () => {
    if (newKeyValue.key && newKeyValue.value) {
      setFormData({
        ...formData,
        data: { ...formData.data, [newKeyValue.key]: newKeyValue.value }
      })
      setNewKeyValue({ key: '', value: '' })
    }
  }

  const removeKeyValue = (key) => {
    const newData = { ...formData.data }
    delete newData[key]
    setFormData({ ...formData, data: newData })
  }

  return (
    <div className="app">
      <header className="app-header">
        <h1>🔐 Simple Vault</h1>
        <p>Manage and store your secrets securely</p>
      </header>

      {error && (
        <div className="error-banner">
          <span>⚠️ {error}</span>
          <button onClick={() => setError(null)}>×</button>
        </div>
      )}

      <div className="toolbar">
        <button className="btn-primary" onClick={() => setShowCreateModal(true)}>
          + Create Secret
        </button>
        <button className="btn-secondary" onClick={fetchSecrets}>
          🔄 Refresh
        </button>
      </div>

      {loading ? (
        <div className="loading">Loading secrets...</div>
      ) : secrets.length === 0 ? (
        <div className="empty-state">
          <p>No secrets found. Create your first secret to get started.</p>
        </div>
      ) : (
        <div className="secrets-grid">
          {secrets.map((secret) => (
            <div key={secret.id} className="secret-card">
              <div className="secret-header">
                <h3>{secret.name}</h3>
                <div className="secret-actions">
                  <button className="btn-icon" onClick={() => openEditModal(secret)} title="Edit">
                    ✏️
                  </button>
                  <button className="btn-icon danger" onClick={() => handleDelete(secret.id)} title="Delete">
                    🗑️
                  </button>
                </div>
              </div>
              {secret.description && <p className="secret-description">{secret.description}</p>}
              <div className="secret-data">
                <strong>Keys:</strong> {Object.keys(secret.data).length}
                <div className="secret-keys">
                  {Object.keys(secret.data).map((key) => (
                    <span key={key} className="key-badge">{key}</span>
                  ))}
                </div>
              </div>
              <div className="secret-meta">
                <small>Created: {new Date(secret.createdAt).toLocaleString()}</small>
              </div>
            </div>
          ))}
        </div>
      )}

      {/* Create Modal */}
      {showCreateModal && (
        <div className="modal-overlay" onClick={() => setShowCreateModal(false)}>
          <div className="modal" onClick={(e) => e.stopPropagation()}>
            <div className="modal-header">
              <h2>Create Secret</h2>
              <button className="btn-close" onClick={() => setShowCreateModal(false)}>×</button>
            </div>
            <form onSubmit={handleCreate} className="modal-body">
              <div className="form-group">
                <label>Name *</label>
                <input
                  type="text"
                  value={formData.name}
                  onChange={(e) => setFormData({ ...formData, name: e.target.value })}
                  required
                />
              </div>
              <div className="form-group">
                <label>Description</label>
                <textarea
                  value={formData.description}
                  onChange={(e) => setFormData({ ...formData, description: e.target.value })}
                  rows="3"
                />
              </div>
              <div className="form-group">
                <label>Key-Value Pairs *</label>
                <div className="key-value-input">
                  <input
                    type="text"
                    placeholder="Key"
                    value={newKeyValue.key}
                    onChange={(e) => setNewKeyValue({ ...newKeyValue, key: e.target.value })}
                  />
                  <input
                    type="password"
                    placeholder="Value"
                    value={newKeyValue.value}
                    onChange={(e) => setNewKeyValue({ ...newKeyValue, value: e.target.value })}
                  />
                  <button type="button" onClick={addKeyValue}>Add</button>
                </div>
                <div className="key-value-list">
                  {Object.entries(formData.data).map(([key, value]) => (
                    <div key={key} className="key-value-item">
                      <span className="key">{key}</span>
                      <span className="value">••••••••</span>
                      <button type="button" onClick={() => removeKeyValue(key)}>Remove</button>
                    </div>
                  ))}
                </div>
              </div>
              <div className="modal-footer">
                <button type="button" className="btn-secondary" onClick={() => setShowCreateModal(false)}>
                  Cancel
                </button>
                <button type="submit" className="btn-primary">Create</button>
              </div>
            </form>
          </div>
        </div>
      )}

      {/* Edit Modal */}
      {showEditModal && (
        <div className="modal-overlay" onClick={() => setShowEditModal(false)}>
          <div className="modal" onClick={(e) => e.stopPropagation()}>
            <div className="modal-header">
              <h2>Edit Secret</h2>
              <button className="btn-close" onClick={() => setShowEditModal(false)}>×</button>
            </div>
            <form onSubmit={handleUpdate} className="modal-body">
              <div className="form-group">
                <label>Name *</label>
                <input
                  type="text"
                  value={formData.name}
                  onChange={(e) => setFormData({ ...formData, name: e.target.value })}
                  required
                />
              </div>
              <div className="form-group">
                <label>Description</label>
                <textarea
                  value={formData.description}
                  onChange={(e) => setFormData({ ...formData, description: e.target.value })}
                  rows="3"
                />
              </div>
              <div className="form-group">
                <label>Key-Value Pairs *</label>
                <div className="key-value-input">
                  <input
                    type="text"
                    placeholder="Key"
                    value={newKeyValue.key}
                    onChange={(e) => setNewKeyValue({ ...newKeyValue, key: e.target.value })}
                  />
                  <input
                    type="password"
                    placeholder="Value"
                    value={newKeyValue.value}
                    onChange={(e) => setNewKeyValue({ ...newKeyValue, value: e.target.value })}
                  />
                  <button type="button" onClick={addKeyValue}>Add</button>
                </div>
                <div className="key-value-list">
                  {Object.entries(formData.data).map(([key, value]) => (
                    <div key={key} className="key-value-item">
                      <span className="key">{key}</span>
                      <span className="value">••••••••</span>
                      <button type="button" onClick={() => removeKeyValue(key)}>Remove</button>
                    </div>
                  ))}
                </div>
              </div>
              <div className="modal-footer">
                <button type="button" className="btn-secondary" onClick={() => setShowEditModal(false)}>
                  Cancel
                </button>
                <button type="submit" className="btn-primary">Update</button>
              </div>
            </form>
          </div>
        </div>
      )}

    </div>
  )
}

export default App
