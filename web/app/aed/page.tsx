"use client";

import { useEffect, useState, Suspense } from "react";
import { Button, Card, CardBody, CardHeader } from "@heroui/react";
import { useSearchParams } from "next/navigation";
import { useTranslation } from "react-i18next";
import "@/lib/i18n/client";
import LanguageSelector from "@/components/shared/LanguageSelector";

// Use environment variable for API URL
const API_URL = process.env.NEXT_PUBLIC_API_URL || "/api";
// For public endpoints, we might need to adjust the path if the backend structure is different
// But usually /public is under the API root. 
// If NEXT_PUBLIC_API_URL is "http://localhost:8080", then we access "http://localhost:8080/public/..."
// If NEXT_PUBLIC_API_URL is "/api", then we access "/api/public/..."
const PUBLIC_API_URL = `${API_URL}/public`;

interface AEDData {
    name: string;
    brand: string;
    location: string;
    data: string;
}

function AEDContent() {
  const searchParams = useSearchParams();
  const uuid = searchParams.get("uuid");
  const { t, i18n } = useTranslation();
  
  const [aedData, setAEDData] = useState<AEDData | null>(null);
  const [loading, setLoading] = useState(true);
  const [error, setError] = useState("");
  const [actionLoading, setActionLoading] = useState(false);
  const [message, setMessage] = useState("");

  useEffect(() => {
    const fetchAED = async () => {
      try {
        const response = await fetch(`${PUBLIC_API_URL}/aed/${uuid}`);
        if (!response.ok) {
            throw new Error("AED not found or disabled");
        }
        const data = await response.json();
        setAEDData(data);
      } catch (err: any) {
        setError(err.message);
      } finally {
        setLoading(false);
      }
    };

    if (uuid) {
        fetchAED();
    }
  }, [uuid]);

  const handleOpen = async () => {
    setActionLoading(true);
    setMessage("");
    try {
        const response = await fetch(`${PUBLIC_API_URL}/aed/${uuid}/open`, {
            method: "POST"
        });
        if (response.ok) {
            setMessage("Command sent successfully!");
        } else {
            try {
                const errorData = await response.json();
                setMessage(errorData.error || `Error ${response.status}`);
            } catch {
                setMessage("Failed to send command.");
            }
        }
    } catch (err) {
        setMessage("Error sending command.");
    } finally {
        setActionLoading(false);
    }
  };

  if (loading) {
    return (
      <div className="flex items-center justify-center min-h-screen bg-background">
        <div className="animate-spin rounded-full h-12 w-12 border-t-2 border-b-2 border-primary"></div>
      </div>
    );
  }

  if (error) {
    return (
      <div className="flex items-center justify-center min-h-screen bg-background p-4">
        <Card className="w-full max-w-md">
            <CardBody className="text-center text-danger">
                <h2 className="text-xl font-bold">Error</h2>
                <p>{error}</p>
            </CardBody>
        </Card>
      </div>
    );
  }

  return (
    <div className="min-h-screen bg-background p-4 flex flex-col items-center justify-center relative text-foreground">
      {/* Language Switcher */}
      <div className="absolute top-4 right-4">
        <LanguageSelector variant="light" className="text-foreground" />
      </div>

      <Card className="w-full max-w-md">
        <CardHeader className="flex flex-col gap-2 items-center pb-6">
            <h1 className="text-3xl font-bold text-primary">AED Control</h1>
            <p className="text-default-500">{aedData?.brand} - {aedData?.name}</p>
        </CardHeader>
        <CardBody className="flex flex-col gap-8 items-center">
            <div className="text-center w-full bg-content2 p-6 rounded-xl">
                <p className="text-sm uppercase font-bold text-default-500 mb-2">Device Location</p>
                <p className="text-2xl font-bold">{aedData?.location ? t(`locations.${aedData.location}`, { defaultValue: aedData.location }) : "Unknown"}</p>
            </div>

            <Button 
                color="success" 
                size="lg" 
                className="w-full h-20 text-2xl font-bold shadow-lg uppercase tracking-wider"
                onPress={handleOpen}
                isLoading={actionLoading}
            >
                {t("common.open", { defaultValue: "OPEN" })}
            </Button>

            {message && (
                <p className={`text-center ${message.includes("success") ? "text-success" : "text-danger"}`}>
                    {message}
                </p>
            )}
        </CardBody>
      </Card>
    </div>
  );
}

export default function AEDPage() {
  return (
    <Suspense fallback={
      <div className="flex items-center justify-center min-h-screen bg-background">
        <div className="animate-spin rounded-full h-12 w-12 border-t-2 border-b-2 border-primary"></div>
      </div>
    }>
      <AEDContent />
    </Suspense>
  );
}
