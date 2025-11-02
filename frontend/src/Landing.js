import React, { useState } from 'react'


export default function Landing({ setUser }) {
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