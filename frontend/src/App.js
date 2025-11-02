import React, { useEffect, useState } from 'react'

function Landing({ setUser }) {
  const [isRegister, setIsRegister] = useState(false)
  const [username, setUsername] = useState("")
  const [email, setEmail] = useState("")
  const [password, setPassword] = useState("")
  const [confirmPassword, setConfirmPassword] = useState("")
  const [err, setErr] = useState(null)

  async function submit(e) {
    e.preventDefault()
    setErr(null)

    // Frontend validation
    if ((!isRegister && !username) || !password || (isRegister && !confirmPassword)) {
      setErr("All required fields must be filled")
      return
    }

    if (isRegister && password !== confirmPassword) {
      setErr("Passwords do not match")
      return
    }

    const url = isRegister ? '/api/register' : '/api/login'
    const body = isRegister ? { email, password, username } : { username, password }
    const res = await fetch(url, {
      method: 'POST',
      headers: { 'Content-Type': 'application/json' },
      credentials: 'include',
      body: JSON.stringify(body)
    })
    if (!res.ok) {
      const text = await res.text()
      setErr(text)
      return
    }
    const u = await res.json()
    setUser(u)
  }

  return (
    <div className="container">
      <h2>{isRegister ? 'Create an account' : 'Login'}</h2>
      <form className="form" onSubmit={submit}>
        <input
          placeholder="Username"
          value={username}
          onChange={e => setUsername(e.target.value)}
          required
        />
        <input
          placeholder="Password"
          type="password"
          value={password}
          onChange={e => setPassword(e.target.value)}
          required
        />
        {isRegister && (
          <>
            <input
              placeholder="Confirm Password"
              type="password"
              value={confirmPassword}
              onChange={e => setConfirmPassword(e.target.value)}
              required
            />
            <input
              placeholder="Email"
              type="email"
              value={email}
              onChange={e => setEmail(e.target.value)}
              required
            />

          </>
        )}
        <div>
          <button className="btn" type="submit">{isRegister ? 'Register' : 'Login'}</button>
          <button
            className="btn"
            type="button"
            onClick={() => {
              setIsRegister(!isRegister)
              setErr(null)
              setPassword("")
              setConfirmPassword("")
            }}
            style={{ marginLeft: 8 }}
          >
            {isRegister ? 'Back to login' : 'New user? Register here'}
          </button>
        </div>
        {err && <p style={{ color: 'red' }}>{err}</p>}
      </form>
    </div>
  )
}

function Navbar({ user, onLogout }) {
  return (
    <div className="nav">
      <div>potbot</div>
      <div>
        <span style={{ marginRight: 12 }}>Hello{user && user.username ? ' ' + user.username : ''}</span>
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
        <h3>Welcome — this is a blank page.</h3>
        <p>You're logged in as: {user.username}</p>
      </div>
    </div>
  )
}

export default App
