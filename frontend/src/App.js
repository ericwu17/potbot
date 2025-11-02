import React, { useEffect, useState } from 'react'
import Landing from './Landing'


function Navbar({ user, onLogout }) {
  return (
    <div className="nav">
      <div>potbot</div>
      <div>
        <span style={{ marginRight: 12 }}></span>
        <a href="#" onClick={(e) => { e.preventDefault(); console.log("Add a new plant not implemented") }}>Add a new plant</a>
        <span style={{ marginRight: 12 }}></span>
        <a href="#" onClick={(e) => { e.preventDefault(); onLogout() }}>Logout</a>
      </div>
    </div>
  )
}

function App() {
  // user state and a wrapper setter that persists to localStorage
  const [user, setUserState] = useState(null)
  const [loading, setLoading] = useState(true)

  function setUser(u) {
    // update React state and persist minimal user info locally
    setUserState(u)
    try {
      if (u) {
        localStorage.setItem('potbot_user', JSON.stringify(u))
      } else {
        localStorage.removeItem('potbot_user')
      }
    } catch (e) {
      // ignore storage errors
      console.warn('localStorage error', e)
    }
  }

  useEffect(() => {
    // hydrate from localStorage first so UI feels instantaneous on refresh
    try {
      const cached = localStorage.getItem('potbot_user')
      if (cached) {
        setUser(JSON.parse(cached))
      }
    } catch (e) {
      // ignore
    }

    // then verify the session with the server and correct state if needed
    fetch('/api/me', { credentials: 'include' })
      .then(res => {
        if (!res.ok) {
          throw new Error('no session')
        }
        
        
        return res.json()
      })
      .then(u => {
          console.log("verified that the user is user " + u)
       })
      .catch(err => {
        // server says unauthenticated -> clear any cached user
        console.log("server says unauthenticated " + err)
        setUser(null)
      })
      .finally(() => setLoading(false))
  }, [])

  async function logout() {
    await fetch('/api/logout', { method: 'POST', credentials: 'include' })
    setUser(null)
  }

  if (loading) return <div className="container">Loading...</div>

  if (!user) return <Landing setUser={setUser} />

  return (
    <div style={{ height: '100%' }}>
      <Navbar user={user} onLogout={logout} />
      <div className="container">
        <h3>Welcome — this is a blank page, since you don't have any plants yet.</h3>
        <p>You're logged in as: {user.username}</p>
      </div>
    </div>
  )
}

export default App
