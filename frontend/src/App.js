import React, {useEffect, useState} from 'react'

function Landing({setUser}){
  const [isRegister, setIsRegister] = useState(false)
  const [email, setEmail] = useState("")
  const [password, setPassword] = useState("")
  const [username, setUsername] = useState("")
  const [err, setErr] = useState(null)

  async function submit(e){
    e.preventDefault()
    setErr(null)
    const url = isRegister ? '/api/register' : '/api/login'
    const body = isRegister ? { email, password, username } : { email, password }
    const res = await fetch(url, {
      method: 'POST',
      headers: {'Content-Type': 'application/json'},
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
        <input placeholder="Email" value={email} onChange={e=>setEmail(e.target.value)} />
        <input placeholder="Password" type="password" value={password} onChange={e=>setPassword(e.target.value)} />
        {isRegister && <input placeholder="Username (optional)" value={username} onChange={e=>setUsername(e.target.value)} />}
        <div>
          <button className="btn" type="submit">{isRegister ? 'Register' : 'Login'}</button>
          <button className="btn" type="button" onClick={()=>setIsRegister(!isRegister)} style={{marginLeft:8}}>{isRegister? 'Switch to login' : 'Switch to register'}</button>
        </div>
        {err && <p style={{color:'red'}}>{err}</p>}
      </form>
    </div>
  )
}

function Navbar({user, onLogout}){
  return (
    <div className="nav">
      <div>potbot</div>
      <div>
        <span style={{marginRight:12}}>Hello{user && user.username ? ' '+ user.username : ''}</span>
        <a href="#" onClick={(e)=>{e.preventDefault(); onLogout()}}>Logout</a>
      </div>
    </div>
  )
}

function App(){
  const [user, setUser] = useState(null)
  const [loading, setLoading] = useState(true)

  useEffect(()=>{
    // check session
    fetch('/api/me', { credentials: 'include' })
      .then(res=>{
        if (!res.ok) throw new Error('no session')
        return res.json()
      })
      .then(u=> setUser(u) )
      .catch(()=>{})
      .finally(()=> setLoading(false))
  },[])

  async function logout(){
    await fetch('/api/logout', { method: 'POST', credentials:'include' })
    setUser(null)
  }

  if (loading) return <div className="container">Loading...</div>

  if (!user) return <Landing setUser={setUser} />

  return (
    <div style={{height:'100%'}}>
      <Navbar user={user} onLogout={logout} />
      <div className="container">
        <h3>Welcome — this is a blank page.</h3>
        <p>You're logged in as: {user.email}</p>
      </div>
    </div>
  )
}

export default App
