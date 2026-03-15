export default function LoginPage() {
  return (
    <div style={{ display: 'flex', flexDirection: 'column', alignItems: 'center', justifyContent: 'center', minHeight: '100vh', gap: 16 }}>
      <h1>Playlist Manager</h1>
      <p>Connect your music accounts to manage playlists in one place.</p>

      <a href="/auth/google/login" style={btnStyle('#FF0000')}>
        🎬 Login with YouTube
      </a>

      <a href="/auth/spotify/login" style={btnStyle('#1DB954')}>
        🎵 Login with Spotify
      </a>
    </div>
  )
}

function btnStyle(bg) {
  return {
    padding: '12px 32px',
    background: bg,
    color: '#fff',
    borderRadius: 8,
    textDecoration: 'none',
    fontWeight: 'bold',
    fontSize: 16,
  }
}
