import { useAuth } from '../context/AuthContext'

export default function Dashboard() {
  const { user, logout } = useAuth()

  return (
    <div style={{ padding: 32 }}>
      <h1>Dashboard</h1>
      <p>Welcome, {user?.display_name ?? 'user'}!</p>

      <section style={{ marginTop: 24 }}>
        <h2>Connected Accounts</h2>
        {/* TODO: show which providers are connected + "Connect Spotify/YouTube" button */}
      </section>

      <section style={{ marginTop: 24 }}>
        <h2>Your Playlists</h2>
        {/* TODO: list playlists from both providers */}
      </section>

      <button onClick={logout} style={{ marginTop: 32 }}>Logout</button>
    </div>
  )
}
