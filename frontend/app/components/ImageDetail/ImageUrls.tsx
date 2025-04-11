'use client'

import { useState } from 'react'
import { motion } from 'framer-motion'
import { ImageData, CopyStatus } from '../../types/image'
import { getFullUrl } from '../../utils/baseUrl'
import { copyToClipboard } from '../../utils/clipboard'

interface ImageUrlsProps {
  image: ImageData
}

interface UrlItemProps {
  title: string
  url: string
  icon: JSX.Element
  iconColor: string
  copyType: string
  copyStatus: CopyStatus | null
  onCopy: (text: string, type: string, e?: React.MouseEvent<HTMLButtonElement>) => void
}

function UrlItem({ title, url, icon, iconColor, copyType, copyStatus, onCopy }: UrlItemProps) {
  return (
    <motion.div
      initial={{ opacity: 0, y: 10 }}
      animate={{ opacity: 1, y: 0 }}
      transition={{ delay: 0.2 }}
    >
      <div className="flex items-center gap-2 mb-2">
        <div className={iconColor}>{icon}</div>
        <div className="font-medium">{title}</div>
      </div>

      <div className="rounded-lg bg-slate-100 dark:bg-slate-900 flex items-center group relative hover:bg-slate-200 dark:hover:bg-slate-800/80 transition-colors duration-200">
        <div className="flex-1 px-4 py-3 text-sm font-mono overflow-hidden text-ellipsis">
          {url}
        </div>
        <button
          onClick={(e) => onCopy(url, copyType, e)}
          className="p-3 hover:bg-slate-200 dark:hover:bg-slate-800 rounded-r-lg transition-colors duration-200 relative"
          title="复制链接"
        >
          {copyStatus && copyStatus.type === copyType ? (
            <svg xmlns="http://www.w3.org/2000/svg" className="h-5 w-5 text-green-500" fill="none" viewBox="0 0 24 24" stroke="currentColor">
              <path strokeLinecap="round" strokeLinejoin="round" strokeWidth="2" d="M5 13l4 4L19 7" />
            </svg>
          ) : (
            <svg xmlns="http://www.w3.org/2000/svg" className="h-5 w-5" fill="none" viewBox="0 0 24 24" stroke="currentColor">
              <path strokeLinecap="round" strokeLinejoin="round" strokeWidth="2" d="M8 16H6a2 2 0 01-2-2V6a2 2 0 012-2h8a2 2 0 012 2v2m-6 12h8a2 2 0 002-2v-8a2 2 0 00-2-2h-8a2 2 0 00-2 2v8a2 2 0 002 2z" />
            </svg>
          )}
          {copyStatus && copyStatus.type === copyType && (
            <span className="absolute -top-8 right-0 bg-black/70 text-white text-xs rounded px-2 py-1">
              已复制!
            </span>
          )}
        </button>
      </div>
    </motion.div>
  )
}

export function ImageUrls({ image }: ImageUrlsProps) {
  const [copyStatus, setCopyStatus] = useState<CopyStatus | null>(null)

  const handleCopy = (text: string, type: string, e?: React.MouseEvent<HTMLButtonElement>) => {
    if (e) e.stopPropagation()

    copyToClipboard(text)
      .then(success => {
        if (success) {
          setCopyStatus({ type })
          setTimeout(() => {
            setCopyStatus(null)
          }, 2000)
        } else {
          console.error("复制失败")
        }
      })
      .catch(err => {
        console.error("复制失败:", err)
      })
  }

  const originalUrl = getFullUrl(image.urls?.original || '')
  const webpUrl = getFullUrl(image.urls?.webp || '')
  const avifUrl = getFullUrl(image.urls?.avif || '')
  const autoUrl = getFullUrl(image.urls?.auto || '')

  return (
    <div className="space-y-6">
      <UrlItem
        title="原始图片"
        url={originalUrl}
        icon={
          <svg xmlns="http://www.w3.org/2000/svg" className="h-5 w-5" fill="none" viewBox="0 0 24 24" stroke="currentColor">
            <path strokeLinecap="round" strokeLinejoin="round" strokeWidth="2" d="M4 16l4.586-4.586a2 2 0 012.828 0L16 16m-2-2l1.586-1.586a2 2 0 012.828 0L20 14" />
          </svg>
        }
        iconColor="text-blue-500"
        copyType="original"
        copyStatus={copyStatus}
        onCopy={handleCopy}
      />

      {autoUrl && (
        <UrlItem
          title="自适应格式"
          url={autoUrl}
          icon={
            <svg xmlns="http://www.w3.org/2000/svg" className="h-5 w-5" fill="none" viewBox="0 0 24 24" stroke="currentColor">
              <path strokeLinecap="round" strokeLinejoin="round" strokeWidth="2" d="M9 12h6m-6 4h6m2 5H7a2 2 0 01-2-2V5a2 2 0 012-2h5.586a1 1 0 01.707.293l5.414 5.414a1 1 0 01.293.707V19a2 2 0 01-2 2z" />
            </svg>
          }
          iconColor="text-amber-500"
          copyType="auto"
          copyStatus={copyStatus}
          onCopy={handleCopy}
        />
      )}

      {webpUrl && (
        <UrlItem
          title="WebP 格式"
          url={webpUrl}
          icon={
            <svg xmlns="http://www.w3.org/2000/svg" className="h-5 w-5" fill="none" viewBox="0 0 24 24" stroke="currentColor">
              <path strokeLinecap="round" strokeLinejoin="round" strokeWidth="2" d="M9 12h6m-6 4h6m2 5H7a2 2 0 01-2-2V5a2 2 0 012-2h5.586a1 1 0 01.707.293l5.414 5.414a1 1 0 01.293.707V19a2 2 0 01-2 2z" />
            </svg>
          }
          iconColor="text-purple-500"
          copyType="webp"
          copyStatus={copyStatus}
          onCopy={handleCopy}
        />
      )}

      {avifUrl && (
        <UrlItem
          title="AVIF 格式"
          url={avifUrl}
          icon={
            <svg xmlns="http://www.w3.org/2000/svg" className="h-5 w-5" fill="none" viewBox="0 0 24 24" stroke="currentColor">
              <path strokeLinecap="round" strokeLinejoin="round" strokeWidth="2" d="M9 12h6m-6 4h6m2 5H7a2 2 0 01-2-2V5a2 2 0 012-2h5.586a1 1 0 01.707.293l5.414 5.414a1 1 0 01.293.707V19a2 2 0 01-2 2z" />
            </svg>
          }
          iconColor="text-green-500"
          copyType="avif"
          copyStatus={copyStatus}
          onCopy={handleCopy}
        />
      )}

      <UrlItem
        title="Markdown 格式"
        url={`![${image.filename}](${originalUrl})`}
        icon={
          <svg xmlns="http://www.w3.org/2000/svg" className="h-5 w-5" fill="none" viewBox="0 0 24 24" stroke="currentColor">
            <path strokeLinecap="round" strokeLinejoin="round" strokeWidth="2" d="M7 8h10M7 12h4m1 8l-4-4H5a2 2 0 01-2-2V6a2 2 0 012-2h14a2 2 0 012 2v8a2 2 0 01-2 2h-3l-4 4z" />
          </svg>
        }
        iconColor="text-amber-500"
        copyType="markdown"
        copyStatus={copyStatus}
        onCopy={handleCopy}
      />
    </div>
  )
} 