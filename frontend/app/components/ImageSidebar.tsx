"use client";

import { useState } from "react";
import Image from "next/image";
import { motion, AnimatePresence } from "framer-motion";
import ImageDetailModal from "./ImageDetail/ImageDetailModal";
import { getFullUrl } from "../utils/baseUrl";

interface ImageSidebarProps {
  isOpen: boolean;
  results: Array<{
    filename: string;
    status: "success" | "error";
    message: string;
    format?: string;
    orientation?: string;
    expiryTime?: string;
    urls?: {
      original: string;
      auto: string;
      webp: string;
      avif: string;
    };
    id?: string;
    path?: string;
  }>;
  onClose: () => void;
  onDelete?: (id: string) => Promise<void>;
}

export default function ImageSidebar({
  isOpen,
  results,
  onClose,
  onDelete,
}: ImageSidebarProps) {
  const [selectedImage, setSelectedImage] = useState<{
    filename: string;
    status: "success" | "error";
    message?: string;
    format?: string;
    orientation?: string;
    expiryTime?: string;
    urls?: {
      original: string;
      auto: string;
      webp: string;
      avif: string;
    };
    id?: string;
    path?: string;
  } | null>(null);
  const [showModal, setShowModal] = useState(false);
  const [tab, setTab] = useState<"all" | "success" | "error">("all");

  const successResults = results.filter(
    (result) => result.status === "success"
  );
  const errorResults = results.filter((result) => result.status === "error");

  // 根据当前标签确定要显示的结果
  const displayResults =
    tab === "all" ? results : tab === "success" ? successResults : errorResults;

  const handleImageClick = (image: any) => {
    setSelectedImage(image);
    setShowModal(true);
  };

  const handleCloseModal = () => {
    setShowModal(false);
  };

  return (
    <>
      <AnimatePresence>
        {isOpen && (
          <motion.div
            initial={{ x: "100%" }}
            animate={{ x: 0 }}
            exit={{ x: "100%" }}
            transition={{ type: "spring", damping: 30, stiffness: 300 }}
            className="fixed top-0 right-0 w-full sm:w-96 h-full bg-white dark:bg-slate-900 shadow-xl z-30 border-l border-slate-200 dark:border-slate-700 overflow-hidden flex flex-col"
          >
            {/* 侧边栏头部 */}
            <div className="flex items-center justify-between p-4 border-b border-slate-200 dark:border-slate-700 bg-gradient-to-r from-indigo-500 to-purple-600 text-white">
              <h2 className="text-lg font-semibold flex items-center">
                <svg
                  xmlns="http://www.w3.org/2000/svg"
                  className="h-5 w-5 mr-2 text-white opacity-90"
                  viewBox="0 0 20 20"
                  fill="currentColor"
                >
                  <path
                    fillRule="evenodd"
                    d="M4 3a2 2 0 00-2 2v10a2 2 0 002 2h12a2 2 0 002-2V5a2 2 0 00-2-2H4zm12 12H4l4-8 3 6 2-4 3 6z"
                    clipRule="evenodd"
                  />
                </svg>
                媒体库 ({results.length})
              </h2>
              <button
                onClick={onClose}
                className="p-2 rounded-full hover:bg-white/10 transition-colors"
              >
                <svg
                  xmlns="http://www.w3.org/2000/svg"
                  className="h-5 w-5"
                  fill="none"
                  viewBox="0 0 24 24"
                  stroke="currentColor"
                >
                  <path
                    strokeLinecap="round"
                    strokeLinejoin="round"
                    strokeWidth="2"
                    d="M6 18L18 6M6 6l12 12"
                  />
                </svg>
              </button>
            </div>

            {/* 标签切换 */}
            <div className="flex border-b border-slate-200 dark:border-slate-700">
              <button
                onClick={() => setTab("all")}
                className={`flex-1 py-3 px-4 text-sm font-medium transition-colors relative ${
                  tab === "all"
                    ? "text-indigo-600 dark:text-indigo-400"
                    : "text-slate-600 dark:text-slate-400 hover:text-slate-900 dark:hover:text-slate-200"
                }`}
              >
                全部 ({results.length})
                {tab === "all" && (
                  <motion.div
                    layoutId="tab-indicator"
                    className="absolute bottom-0 left-0 right-0 h-0.5 bg-indigo-600 dark:bg-indigo-400"
                  />
                )}
              </button>
              <button
                onClick={() => setTab("success")}
                className={`flex-1 py-3 px-4 text-sm font-medium transition-colors relative ${
                  tab === "success"
                    ? "text-green-600 dark:text-green-400"
                    : "text-slate-600 dark:text-slate-400 hover:text-slate-900 dark:hover:text-slate-200"
                }`}
              >
                成功 ({successResults.length})
                {tab === "success" && (
                  <motion.div
                    layoutId="tab-indicator"
                    className="absolute bottom-0 left-0 right-0 h-0.5 bg-green-600 dark:bg-green-400"
                  />
                )}
              </button>
              <button
                onClick={() => setTab("error")}
                className={`flex-1 py-3 px-4 text-sm font-medium transition-colors relative ${
                  tab === "error"
                    ? "text-red-600 dark:text-red-400"
                    : "text-slate-600 dark:text-slate-400 hover:text-slate-900 dark:hover:text-slate-200"
                }`}
              >
                失败 ({errorResults.length})
                {tab === "error" && (
                  <motion.div
                    layoutId="tab-indicator"
                    className="absolute bottom-0 left-0 right-0 h-0.5 bg-red-600 dark:bg-red-400"
                  />
                )}
              </button>
            </div>

            {/* 侧边栏内容 */}
            <div className="flex-1 overflow-y-auto p-4">
              {displayResults.length === 0 ? (
                <div className="flex flex-col items-center justify-center h-full text-center text-slate-500 dark:text-slate-400 p-6">
                  <svg
                    xmlns="http://www.w3.org/2000/svg"
                    className="h-16 w-16 mb-4 text-slate-300 dark:text-slate-600"
                    fill="none"
                    viewBox="0 0 24 24"
                    stroke="currentColor"
                  >
                    <path
                      strokeLinecap="round"
                      strokeLinejoin="round"
                      strokeWidth="2"
                      d="M4 16l4.586-4.586a2 2 0 012.828 0L16 16m-2-2l1.586-1.586a2 2 0 012.828 0L20 14m-6-6h.01M6 20h12a2 2 0 002-2V6a2 2 0 00-2-2H6a2 2 0 00-2 2v12a2 2 0 002 2z"
                    />
                  </svg>
                  <p className="text-lg font-medium mb-2">暂无图片</p>
                  <p className="text-sm">
                    {tab === "all"
                      ? "上传完成的图片将会显示在这里"
                      : tab === "success"
                      ? "没有成功上传的图片"
                      : "没有上传失败的图片"}
                  </p>
                </div>
              ) : (
                <div className="space-y-4">
                  {/* 显示选定的结果 */}
                  <div className="grid grid-cols-2 gap-3">
                    {displayResults.map((result, index) => (
                      <motion.div
                        key={result.id || `${tab}-${index}`}
                        initial={{ opacity: 0, scale: 0.9 }}
                        animate={{ opacity: 1, scale: 1 }}
                        transition={{ delay: index * 0.05 }}
                        className={`relative rounded-lg overflow-hidden border ${
                          result.status === "success"
                            ? "border-slate-200 dark:border-slate-700 bg-white dark:bg-slate-800"
                            : "border-red-200 dark:border-red-800 bg-red-50 dark:bg-red-900/20"
                        } shadow-sm hover:shadow-md transition-all cursor-pointer group`}
                        onClick={() =>
                          result.status === "success" &&
                          handleImageClick(result)
                        }
                      >
                        {result.status === "success" ? (
                          <>
                            <div className="aspect-square relative bg-slate-50 dark:bg-slate-900">
                              {result.urls?.original && (
                                <Image
                                  src={getFullUrl(result.urls.original)}
                                  alt={result.filename}
                                  fill
                                  className="object-cover group-hover:scale-105 transition-transform duration-300"
                                  sizes="(max-width: 768px) 50vw, 33vw"
                                />
                              )}
                              <div className="absolute inset-0 bg-gradient-to-t from-black/60 via-transparent to-transparent opacity-0 group-hover:opacity-100 transition-opacity duration-300"></div>
                              <div className="absolute top-1 right-1">
                                <span className="text-xs px-1.5 py-0.5 bg-green-500/80 text-white rounded-full">
                                  完成
                                </span>
                              </div>
                              <div className="absolute bottom-0 left-0 right-0 p-2 text-white transform translate-y-full group-hover:translate-y-0 transition-transform duration-300">
                                <p
                                  className="text-xs truncate"
                                  title={result.filename}
                                >
                                  {result.filename}
                                </p>
                                {result.expiryTime && (
                                  <p className="text-xs mt-1">
                                    <span className="bg-yellow-500/80 text-white px-1 py-0.5 rounded text-[10px]">
                                      过期时间:{" "}
                                      {new Date(
                                        result.expiryTime
                                      ).toLocaleString()}
                                    </span>
                                  </p>
                                )}
                              </div>
                            </div>
                          </>
                        ) : (
                          <div className="p-3 h-full flex flex-col">
                            <div className="flex items-start space-x-2">
                              <svg
                                xmlns="http://www.w3.org/2000/svg"
                                className="h-5 w-5 text-red-500 flex-shrink-0 mt-0.5"
                                viewBox="0 0 20 20"
                                fill="currentColor"
                              >
                                <path
                                  fillRule="evenodd"
                                  d="M18 10a8 8 0 11-16 0 8 8 0 0116 0zm-7 4a1 1 0 11-2 0 1 1 0 012 0zm-1-9a1 1 0 00-1 1v4a1 1 0 102 0V6a1 1 0 00-1-1z"
                                  clipRule="evenodd"
                                />
                              </svg>
                              <div>
                                <p className="font-medium text-sm text-red-600 dark:text-red-400">
                                  {result.filename}
                                </p>
                                <p className="text-xs text-red-500 dark:text-red-300 mt-1">
                                  {result.message}
                                </p>
                              </div>
                            </div>
                          </div>
                        )}
                      </motion.div>
                    ))}
                  </div>
                </div>
              )}
            </div>
          </motion.div>
        )}
      </AnimatePresence>

      {/* 图片详情模态框 */}
      <ImageDetailModal
        image={selectedImage}
        isOpen={showModal}
        onClose={handleCloseModal}
        onDelete={onDelete}
      />

      {/* 背景遮罩 */}
      <AnimatePresence>
        {isOpen && (
          <motion.div
            initial={{ opacity: 0 }}
            animate={{ opacity: 1 }}
            exit={{ opacity: 0 }}
            className="fixed inset-0 bg-black/20 dark:bg-black/50 backdrop-blur-sm z-20 sm:block hidden"
            onClick={onClose}
          />
        )}
      </AnimatePresence>
    </>
  );
}
