export interface ImageData {
  filename: string
  status: 'success' | 'error'
  message?: string
  format?: string
  orientation?: string
  expiryTime?: string
  urls?: {
    original: string
    webp: string
    avif: string
    auto: string
  }
  id?: string
  path?: string
}

export interface CopyStatus {
  type: string
} 