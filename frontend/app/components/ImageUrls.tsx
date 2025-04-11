import { useState } from "react";
import { ImageFile } from "../types";
import { buildMarkdownLink } from "../utils/imageUtils";
import { copyToClipboard } from "../utils/clipboard";
import { getFullUrl } from "../utils/baseUrl";

interface ImageUrlsProps {
  image: ImageFile;
}

export const ImageUrls = ({ image }: ImageUrlsProps) => {
  const [copyStatus, setCopyStatus] = useState<{ type: string } | null>(null);

  const handleCopy = (
    text: string,
    type: string,
    e?: React.MouseEvent<HTMLButtonElement>
  ) => {
    if (e) e.stopPropagation();

    copyToClipboard(text)
      .then((success) => {
        if (success) {
          setCopyStatus({ type });
          setTimeout(() => {
            setCopyStatus(null);
          }, 2000);
        } else {
          console.error("复制失败");
        }
      })
      .catch((err) => {
        console.error("复制失败:", err);
      });
  };

  const originalUrl = getFullUrl(image.urls?.original || "");
  const autoUrl = getFullUrl(image.urls?.auto || "");
  const webpUrl = getFullUrl(image.urls?.webp || "");
  const avifUrl = getFullUrl(image.urls?.avif || "");
  const currentFormatUrl =
    image.format.toLowerCase() === "webp"
      ? webpUrl
      : image.format.toLowerCase() === "avif"
      ? avifUrl
      : originalUrl;
  const markdownLink = buildMarkdownLink(currentFormatUrl!, image.filename);

  const CopyButton = ({ type, text }: { type: string; text: string }) => (
    <button
      onClick={(e) => handleCopy(text, type, e)}
      className="p-1 hover:bg-gray-200 dark:hover:bg-gray-700 rounded-r-lg transition-colors relative"
      title="复制链接"
    >
      {copyStatus && copyStatus.type === type ? (
        <svg
          xmlns="http://www.w3.org/2000/svg"
          className="h-4 w-4 text-green-500"
          fill="none"
          viewBox="0 0 24 24"
          stroke="currentColor"
        >
          <path
            strokeLinecap="round"
            strokeLinejoin="round"
            strokeWidth="2"
            d="M5 13l4 4L19 7"
          />
        </svg>
      ) : (
        <svg
          xmlns="http://www.w3.org/2000/svg"
          className="h-4 w-4 text-gray-500 dark:text-gray-400"
          fill="none"
          viewBox="0 0 24 24"
          stroke="currentColor"
        >
          <path
            strokeLinecap="round"
            strokeLinejoin="round"
            strokeWidth={2}
            d="M8 16H6a2 2 0 01-2-2V6a2 2 0 012-2h8a2 2 0 012 2v2m-6 12h8a2 2 0 002-2v-8a2 2 0 00-2-2h-8a2 2 0 00-2 2v8a2 2 0 002 2z"
          />
        </svg>
      )}
    </button>
  );

  const UrlBox = ({
    icon,
    label,
    url,
    type,
  }: {
    icon: string;
    label: string;
    url: string;
    type: string;
  }) => (
    <div className="mb-1">
      <div className="flex items-center gap-1 mb-1">
        <svg
          xmlns="http://www.w3.org/2000/svg"
          className={`h-3 w-3 ${icon}`}
          fill="none"
          viewBox="0 0 24 24"
          stroke="currentColor"
        >
          <path
            strokeLinecap="round"
            strokeLinejoin="round"
            strokeWidth="2"
            d="M4 16l4.586-4.586a2 2 0 012.828 0L16 16m-2-2l1.586-1.586a2 2 0 012.828 0L20 14"
          />
        </svg>
        <div className="text-xs font-medium">{label}</div>
      </div>

      <div className="rounded-lg bg-gray-50 dark:bg-gray-800 flex items-center relative">
        <div className="flex-1 px-2 py-1 text-xs font-mono overflow-hidden text-ellipsis">
          {url}
        </div>
        <CopyButton type={type} text={url} />
      </div>
    </div>
  );

  return (
    <div className="space-y-1">
      <div className="mb-1">
        <label className="block text-xs font-medium text-gray-700 dark:text-gray-300 mb-1">
          可用格式
        </label>
      </div>

      {/* 原始格式链接 */}
      <UrlBox
        icon="text-blue-500"
        label={image.format.toLowerCase() === "gif" ? "GIF 动图" : "原始格式"}
        url={originalUrl!}
        type="original"
      />

      {/* 仅在非GIF图片时显示WebP和AVIF格式 */}
      {image.format.toLowerCase() !== "gif" && (
        <>
          <UrlBox
            icon="text-purple-500"
            label="自适应格式（Webp/Avif）"
            url={autoUrl!}
            type="auto"
          />
          <UrlBox
            icon="text-purple-500"
            label="WebP格式"
            url={webpUrl!}
            type="webp"
          />
          <UrlBox
            icon="text-green-500"
            label="AVIF格式"
            url={avifUrl!}
            type="avif"
          />
        </>
      )}

      {/* Markdown格式链接 */}
      <UrlBox
        icon="text-amber-500"
        label="Markdown格式"
        url={markdownLink}
        type="markdown"
      />
    </div>
  );
};
