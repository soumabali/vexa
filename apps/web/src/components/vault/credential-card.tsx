"use client";

import React, { useState } from "react";
import { motion } from "framer-motion";
import { Card, CardContent } from "@/components/ui/card";
import { Badge } from "@/components/ui/badge";
import { Button } from "@/components/ui/button";
import {
  Tooltip,
  TooltipContent,
  TooltipProvider,
  TooltipTrigger,
} from "@/components/ui/tooltip";
import {
  Lock,
  Star,
  Copy,
  Eye,
  EyeOff,
  ExternalLink,
  MoreHorizontal,
  Trash2,
  Edit,
  Share2,
} from "lucide-react";
import type { Credential, CredentialType } from "./credential-list";

interface CredentialCardProps {
  credential: Credential;
  onUpdate: (id: string, data: Partial<Credential>) => void;
  onDelete: (id: string) => void;
  onFavorite: (id: string) => void;
  onClick: () => void;
  viewMode: "grid" | "list";
}

const typeColors: Record<CredentialType, string> = {
  "ssh-key": "bg-blue-100 text-blue-800 dark:bg-blue-900 dark:text-blue-200",
  password: "bg-yellow-100 text-yellow-800 dark:bg-yellow-900 dark:text-yellow-200",
  "api-key": "bg-purple-100 text-purple-800 dark:bg-purple-900 dark:text-purple-200",
  certificate: "bg-green-100 text-green-800 dark:bg-green-900 dark:text-green-200",
  note: "bg-gray-100 text-gray-800 dark:bg-gray-800 dark:text-gray-200",
};

const typeLabels: Record<CredentialType, string> = {
  "ssh-key": "SSH Key",
  password: "Password",
  "api-key": "API Key",
  certificate: "Certificate",
  note: "Secure Note",
};

export function CredentialCard({
  credential,
  onUpdate,
  onDelete,
  onFavorite,
  onClick,
  viewMode,
}: CredentialCardProps) {
  const [showSecret, setShowSecret] = useState(false);
  const [isHovered, setIsHovered] = useState(false);

  const handleCopy = async (text: string) => {
    await navigator.clipboard.writeText(text);
    // Could show toast here
  };

  if (viewMode === "list") {
    return (
      <Card
        className="cursor-pointer hover:shadow-md transition-shadow"
        onClick={onClick}
        onMouseEnter={() => setIsHovered(true)}
        onMouseLeave={() => setIsHovered(false)}
      >
        <CardContent className="p-4 flex items-center gap-4">
          <div className="flex items-center gap-3 flex-1">
            <Badge className={typeColors[credential.type]}>
              {typeLabels[credential.type]}
            </Badge>
            <div className="flex-1">
              <h3 className="font-medium">{credential.name}</h3>
              {credential.host && (
                <p className="text-sm text-muted-foreground">
                  {credential.username ? `${credential.username}@` : ""}
                  {credential.host}
                  {credential.port ? `:${credential.port}` : ""}
                </p>
              )}
            </div>
          </div>

          <div className="flex items-center gap-2">
            {credential.tags.map((tag) => (
              <Badge key={tag} variant="outline" className="text-xs">
                {tag}
              </Badge>
            ))}
          </div>

          <div className="flex items-center gap-1">
            <Button
              variant="ghost"
              size="sm"
              onClick={(e) => {
                e.stopPropagation();
                onFavorite(credential.id);
              }}
            >
              <Star
                className={`h-4 w-4 ${
                  credential.isFavorite ? "fill-yellow-400 text-yellow-400" : ""
                }`}
              />
            </Button>
            <Button
              variant="ghost"
              size="sm"
              onClick={(e) => {
                e.stopPropagation();
                // Share dialog
              }}
            >
              <Share2 className="h-4 w-4" />
            </Button>
            <Button
              variant="ghost"
              size="sm"
              onClick={(e) => {
                e.stopPropagation();
                onDelete(credential.id);
              }}
            >
              <Trash2 className="h-4 w-4 text-destructive" />
            </Button>
          </div>
        </CardContent>
      </Card>
    );
  }

  return (
    <TooltipProvider>
      <motion.div
        whileHover={{ scale: 1.02 }}
        whileTap={{ scale: 0.98 }}
      >
        <Card
          className="cursor-pointer hover:shadow-lg transition-all duration-200 relative overflow-hidden"
          onClick={onClick}
          onMouseEnter={() => setIsHovered(true)}
          onMouseLeave={() => setIsHovered(false)}
        >
          {/* Favorite indicator */}
          {credential.isFavorite && (
            <div className="absolute top-2 right-2">
              <Star className="h-4 w-4 fill-yellow-400 text-yellow-400" />
            </div>
          )}

          <CardContent className="p-4 space-y-3">
            {/* Header */}
            <div className="flex items-start justify-between">
              <div className="space-y-1">
                <Badge className={typeColors[credential.type]}>
                  {typeLabels[credential.type]}
                </Badge>
                <h3 className="font-semibold text-lg truncate">{credential.name}</h3>
              </div>
            </div>

            {/* Host info */}
            {credential.host && (
              <div className="text-sm text-muted-foreground">
                <p>
                  {credential.username ? `${credential.username}@` : ""}
                  {credential.host}
                  {credential.port ? `:${credential.port}` : ""}
                </p>
              </div>
            )}

            {/* Tags */}
            {credential.tags.length > 0 && (
              <div className="flex flex-wrap gap-1">
                {credential.tags.slice(0, 3).map((tag) => (
                  <Badge key={tag} variant="outline" className="text-xs">
                    {tag}
                  </Badge>
                ))}
                {credential.tags.length > 3 && (
                  <Badge variant="outline" className="text-xs">
                    +{credential.tags.length - 3}
                  </Badge>
                )}
              </div>
            )}

            {/* Actions */}
            <div className="flex items-center justify-between pt-2 border-t">
              <div className="flex items-center gap-1">
                <Tooltip>
                  <TooltipTrigger asChild>
                    <Button
                      variant="ghost"
                      size="sm"
                      onClick={(e) => {
                        e.stopPropagation();
                        onFavorite(credential.id);
                      }}
                    >
                      <Star
                        className={`h-4 w-4 ${
                          credential.isFavorite
                            ? "fill-yellow-400 text-yellow-400"
                            : ""
                        }`}
                      />
                    </Button>
                  </TooltipTrigger>
                  <TooltipContent>
                    {credential.isFavorite ? "Remove favorite" : "Add favorite"}
                  </TooltipContent>
                </Tooltip>

                <Tooltip>
                  <TooltipTrigger asChild>
                    <Button
                      variant="ghost"
                      size="sm"
                      onClick={(e) => {
                        e.stopPropagation();
                        // Open share dialog
                      }}
                    >
                      <Share2 className="h-4 w-4" />
                    </Button>
                  </TooltipTrigger>
                  <TooltipContent>Share</TooltipContent>
                </Tooltip>
              </div>

              <div className="flex items-center gap-1">
                <Tooltip>
                  <TooltipTrigger asChild>
                    <Button
                      variant="ghost"
                      size="sm"
                      onClick={(e) => {
                        e.stopPropagation();
                        // Copy connection string
                        handleCopy(
                          `${credential.username}@${credential.host}:${credential.port}`
                        );
                      }}
                    >
                      <Copy className="h-4 w-4" />
                    </Button>
                  </TooltipTrigger>
                  <TooltipContent>Copy connection</TooltipContent>
                </Tooltip>

                <Tooltip>
                  <TooltipTrigger asChild>
                    <Button
                      variant="ghost"
                      size="sm"
                      onClick={(e) => {
                        e.stopPropagation();
                        onDelete(credential.id);
                      }}
                    >
                      <Trash2 className="h-4 w-4 text-destructive" />
                    </Button>
                  </TooltipTrigger>
                  <TooltipContent>Delete</TooltipContent>
                </Tooltip>
              </div>
            </div>
          </CardContent>
        </Card>
      </motion.div>
    </TooltipProvider>
  );
}
