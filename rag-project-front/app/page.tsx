"use client"

import { useState } from "react"
import LoginPage from "@/components/login-page"
import Dashboard from "@/components/dashboard"

export default function Home() {
  const [isAuthenticated, setIsAuthenticated] = useState(false)
  const [username, setUsername] = useState("")
  const [token, setToken] = useState<string | null>(null)

  const handleLogin = (user: string, authToken: string) => {
    setUsername(user)
    setToken(authToken)
    setIsAuthenticated(true)
  }

  const handleLogout = () => {
    setIsAuthenticated(false)
    setUsername("")
    setToken(null)
    if (typeof window !== "undefined") {
      localStorage.removeItem("authToken")
      localStorage.removeItem("authUsername")
    }
  }

  return (
    <main className="min-h-screen">
      {!isAuthenticated ? (
        <LoginPage onLogin={handleLogin} />
      ) : (
        token && <Dashboard username={username} token={token} onLogout={handleLogout} />
      )}
    </main>
  )
}
