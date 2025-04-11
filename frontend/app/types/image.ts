export interface ImageData {
  filename: string
  status: 'success' | 'error'
  message?: string
  format?: string
  orientation?: string
  expiryTime?: string
  urls?: {
    original: string
    auto: string
    webp: string
    avif: string
  }
  id?: string
  path?: string
}

export interface CopyStatus {
  type: string
} 