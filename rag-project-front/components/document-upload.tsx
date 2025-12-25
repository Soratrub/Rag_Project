"use client"

import type React from "react"

import { useState, useRef } from "react"
import { Button } from "@/components/ui/button"
import { Card, CardContent } from "@/components/ui/card"
import { Progress } from "@/components/ui/progress"
import { Upload, File, CheckCircle2, XCircle } from "lucide-react"
import { cn } from "@/lib/utils"

const API_BASE_URL = process.env.NEXT_PUBLIC_API_BASE_URL ?? "http://localhost:3000"

type UploadStatus = "idle" | "uploading" | "success" | "error"

interface DocumentUploadProps {
  token: string
  onDocumentUploaded?: (documentId: number) => void
}

export default function DocumentUpload({ token, onDocumentUploaded }: DocumentUploadProps) {
  const [isDragging, setIsDragging] = useState(false)
  const [uploadStatus, setUploadStatus] = useState<UploadStatus>("idle")
  const [uploadProgress, setUploadProgress] = useState(0)
  const [fileName, setFileName] = useState("")
  const [error, setError] = useState<string | null>(null)
  const fileInputRef = useRef<HTMLInputElement>(null)

  const handleDragOver = (e: React.DragEvent) => {
    e.preventDefault()
    setIsDragging(true)
  }

  const handleDragLeave = () => {
    setIsDragging(false)
  }

  const handleDrop = (e: React.DragEvent) => {
    e.preventDefault()
    setIsDragging(false)
    const files = e.dataTransfer.files
    if (files.length > 0 && files[0].type === "application/pdf") {
      handleFileUpload(files[0])
    }
  }

  const handleFileSelect = (e: React.ChangeEvent<HTMLInputElement>) => {
    const files = e.target.files
    if (files && files.length > 0) {
      handleFileUpload(files[0])
    }
  }

  const handleFileUpload = async (file: File) => {
    setFileName(file.name)
    setUploadStatus("uploading")
    setUploadProgress(0)
    setError(null)

    try {
      const formData = new FormData()
      formData.append("file", file)

      const response = await fetch(`${API_BASE_URL}/api/upload`, {
        method: "POST",
        headers: {
          Authorization: `Bearer ${token}`,
        },
        body: formData,
      })

      if (!response.ok) {
        const data = await response.json().catch(() => ({}))
        const message = (data as { error?: string }).error ?? "Upload failed"
        throw new Error(message)
      }

      const data = (await response.json()) as { document_id?: number; filename?: string }

      setUploadProgress(100)
      setUploadStatus("success")

      if (data.filename) {
        setFileName(data.filename)
      }

      if (onDocumentUploaded && typeof data.document_id === "number") {
        onDocumentUploaded(data.document_id)
      }
    } catch (err) {
      const message = err instanceof Error ? err.message : "Upload failed"
      setError(message)
      setUploadStatus("error")
      setUploadProgress(0)
    }
  }

  const resetUpload = () => {
    setUploadStatus("idle")
    setUploadProgress(0)
    setFileName("")
    setError(null)
    if (fileInputRef.current) {
      fileInputRef.current.value = ""
    }
  }

  return (
    <div className="h-full flex flex-col">
      <div className="border-b border-border bg-card px-6 py-4">
        <h2 className="text-2xl font-bold text-foreground">Document Upload</h2>
        <p className="text-sm text-muted-foreground mt-1">Upload PDF documents to enhance the AI knowledge base</p>
      </div>

      <div className="flex-1 overflow-auto p-6">
        <div className="max-w-2xl mx-auto space-y-6">
          <Card
            className={cn(
              "border-2 border-dashed transition-colors",
              isDragging ? "border-primary bg-primary/5" : "border-border",
              uploadStatus === "idle" && "hover:border-primary/50",
            )}
            onDragOver={handleDragOver}
            onDragLeave={handleDragLeave}
            onDrop={handleDrop}
          >
            <CardContent className="flex flex-col items-center justify-center p-12 text-center">
              {error && (
                <p className="mb-4 text-sm text-red-500">
                  {error}
                </p>
              )}
              {uploadStatus === "idle" && (
                <>
                  <div className="w-16 h-16 rounded-full bg-primary/10 flex items-center justify-center mb-4">
                    <Upload className="w-8 h-8 text-primary" />
                  </div>
                  <h3 className="text-lg font-semibold mb-2">Drop your PDF here</h3>
                  <p className="text-sm text-muted-foreground mb-6">or click the button below to browse files</p>
                  <input ref={fileInputRef} type="file" accept=".pdf" onChange={handleFileSelect} className="hidden" />
                  <Button onClick={() => fileInputRef.current?.click()}>
                    <File className="w-4 h-4 mr-2" />
                    Select PDF File
                  </Button>
                </>
              )}

              {uploadStatus === "uploading" && (
                <div className="w-full max-w-md">
                  <div className="w-16 h-16 rounded-full bg-primary/10 flex items-center justify-center mb-4 mx-auto">
                    <Upload className="w-8 h-8 text-primary animate-pulse" />
                  </div>
                  <h3 className="text-lg font-semibold mb-2">Uploading...</h3>
                  <p className="text-sm text-muted-foreground mb-4">{fileName}</p>
                  <Progress value={uploadProgress} className="mb-2" />
                  <p className="text-xs text-muted-foreground">{uploadProgress}% complete</p>
                </div>
              )}

              {uploadStatus === "success" && (
                <>
                  <div className="w-16 h-16 rounded-full bg-accent/10 flex items-center justify-center mb-4">
                    <CheckCircle2 className="w-8 h-8 text-accent" />
                  </div>
                  <h3 className="text-lg font-semibold mb-2">Upload Successful!</h3>
                  <p className="text-sm text-muted-foreground mb-6">{fileName}</p>
                  <Button onClick={resetUpload}>Upload Another File</Button>
                </>
              )}

              {uploadStatus === "error" && (
                <>
                  <div className="w-16 h-16 rounded-full bg-destructive/10 flex items-center justify-center mb-4">
                    <XCircle className="w-8 h-8 text-destructive" />
                  </div>
                  <h3 className="text-lg font-semibold mb-2">Upload Failed</h3>
                  <p className="text-sm text-muted-foreground mb-6">
                    There was an error uploading your file. Please try again.
                  </p>
                  <Button onClick={resetUpload}>Try Again</Button>
                </>
              )}
            </CardContent>
          </Card>

          <Card className="border-border">
            <CardContent className="p-6">
              <h3 className="font-semibold mb-3">Upload Guidelines</h3>
              <ul className="space-y-2 text-sm text-muted-foreground">
                <li className="flex items-start gap-2">
                  <span className="text-primary mt-0.5">•</span>
                  <span>Only PDF files are accepted for processing</span>
                </li>
                <li className="flex items-start gap-2">
                  <span className="text-primary mt-0.5">•</span>
                  <span>Maximum file size: 50MB</span>
                </li>
                <li className="flex items-start gap-2">
                  <span className="text-primary mt-0.5">•</span>
                  <span>Documents will be processed and added to the knowledge base</span>
                </li>
                <li className="flex items-start gap-2">
                  <span className="text-primary mt-0.5">•</span>
                  <span>Processing time varies based on document size</span>
                </li>
              </ul>
            </CardContent>
          </Card>
        </div>
      </div>
    </div>
  )
}
