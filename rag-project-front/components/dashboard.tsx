"use client"

import { useState } from "react"
import { Button } from "@/components/ui/button"
import { Avatar, AvatarFallback } from "@/components/ui/avatar"
import { FileUp, MessageSquare, LogOut } from "lucide-react"
import DocumentUpload from "@/components/document-upload"
import ChatInterface from "@/components/chat-interface"

interface DashboardProps {
  username: string
  token: string
  onLogout: () => void
}

type Section = "upload" | "chat"

export default function Dashboard({ username, token, onLogout }: DashboardProps) {
  const [activeSection, setActiveSection] = useState<Section>("upload")
  const [activeDocumentId, setActiveDocumentId] = useState<number | null>(null)

  return (
    <div className="flex h-screen bg-background">
      {/* Sidebar */}
      <aside className="w-64 border-r border-border bg-card flex flex-col">
        <div className="p-6 border-b border-border">
          <h1 className="text-xl font-bold text-foreground">RAG Platform</h1>
          <p className="text-sm text-muted-foreground mt-1">Knowledge Assistant</p>
        </div>

        <nav className="flex-1 p-4 space-y-2">
          <Button
            variant={activeSection === "upload" ? "secondary" : "ghost"}
            className="w-full justify-start"
            onClick={() => setActiveSection("upload")}
          >
            <FileUp className="w-4 h-4 mr-3" />
            Document Upload
          </Button>
          <Button
            variant={activeSection === "chat" ? "secondary" : "ghost"}
            className="w-full justify-start"
            onClick={() => setActiveSection("chat")}
          >
            <MessageSquare className="w-4 h-4 mr-3" />
            Chat
          </Button>
        </nav>

        <div className="p-4 border-t border-border">
          <div className="flex items-center gap-3 mb-3">
            <Avatar className="w-8 h-8 bg-primary">
              <AvatarFallback className="bg-primary text-primary-foreground text-sm">
                {username.charAt(0).toUpperCase()}
              </AvatarFallback>
            </Avatar>
            <div className="flex-1 min-w-0">
              <p className="text-sm font-medium text-foreground truncate">{username}</p>
              <p className="text-xs text-muted-foreground">User</p>
            </div>
          </div>
          <Button variant="outline" className="w-full bg-transparent" onClick={onLogout}>
            <LogOut className="w-4 h-4 mr-2" />
            Logout
          </Button>
        </div>
      </aside>

      {/* Main Content */}
      <main className="flex-1 overflow-hidden">
        {activeSection === "upload" && (
          <DocumentUpload
            token={token}
            onDocumentUploaded={(id) => {
              setActiveDocumentId(id)
              setActiveSection("chat")
            }}
          />
        )}
        {activeSection === "chat" && <ChatInterface token={token} documentId={activeDocumentId ?? undefined} />}
      </main>
    </div>
  )
}
