"use client";

import React from "react";

export default function MainLayout({ children }: { children: React.ReactNode }) {
  return (
    <div className="min-h-[calc(100vh-var(--navbar-height))] bg-background text-foreground pb-20">
      <main className="container mx-auto p-4 max-w-4xl h-full">
        {children}
      </main>
    </div>
  );
}
