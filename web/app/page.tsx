"use client";

import { useEffect, useState } from "react";
import { 
  Button,
  Card,
  CardBody,
  useDisclosure,
  Modal,
  ModalContent,
  ModalHeader,
  ModalBody,
  ModalFooter,
  Input,
  Select,
  SelectItem,
  Switch,
} from "@heroui/react";
import { useTranslation } from "react-i18next";
import { FaPlus, FaTrash, FaQrcode } from "react-icons/fa";
import QRCode from "react-qr-code";
import "../lib/i18n/client";
import MainLayout from "@/components/layout/MainLayout";

// Use environment variable for API URL
const BASE_API_URL = process.env.NEXT_PUBLIC_API_URL || "/api";
const API_URL = `${BASE_API_URL}/admin`;
const PUBLIC_API_URL = `${BASE_API_URL}/public`;

interface Vehicle {
    ID: number;
    name: string;
    brand: string;
    location: string;
    data: string;
    enabled: boolean;
    CreatedAt: string;
    UpdatedAt: string;
    qr_code?: {
        uuid: string;
    };
}

interface Brand {
    name: string;
    fields: { [key: string]: string };
}

const LOCATIONS = [
    "Front Trunk", 
    "Rear Trunk", 
    "Front Left", 
    "Front Right", 
    "Rear Left", 
    "Rear Right", 
    "Custom"
];

export default function Dashboard() {
  const { t } = useTranslation();
  
  // Data
  const [vehicles, setVehicles] = useState<Vehicle[]>([]);
  const [brands, setBrands] = useState<Brand[]>([]);
  const [publicDomain, setPublicDomain] = useState("");
  
  // Modal state
  const {isOpen, onOpen, onOpenChange} = useDisclosure();
  const [editingVehicle, setEditingVehicle] = useState<Vehicle | null>(null);
  
  // QR Modal state
  const {isOpen: isQRModalOpen, onOpen: onQROpen, onOpenChange: onQROpenChange} = useDisclosure();
  const [qrCodeData, setQrCodeData] = useState<string>("");
  const [selectedVehicleName, setSelectedVehicleName] = useState<string>("");

  // Form state
  const [formData, setFormData] = useState({
    name: "",
    brand: "",
    locationType: "",
    customLocation: "",
    location: "",
    data: "{}",
    enabled: true
  });
  
  const [apiRegion, setApiRegion] = useState<string>("");
  
  // Tesla Token Automation
  const [teslaTokenMode, setTeslaTokenMode] = useState<"manual" | "auto">("manual");
  const [teslaVerifier, setTeslaVerifier] = useState("");
  const [teslaCallbackUrl, setTeslaCallbackUrl] = useState("");
  const [teslaTokenLoading, setTeslaTokenLoading] = useState(false);
  
  // Step state
  const [currentStep, setCurrentStep] = useState(0); // 0: Basic, 1: Token, 2: Vehicle Selection
  const [teslaProducts, setTeslaProducts] = useState<any[]>([]);
  const [fetchingVehicles, setFetchVehicles] = useState(false);

  // Dynamic form data for brand specific fields
  const [dynamicFormData, setDynamicFormData] = useState<{ [key: string]: string }>({});

  const fetchPublicConfig = async () => {
    try {
        const response = await fetch(`${PUBLIC_API_URL}/config`);
        if (response.ok) {
            const data = await response.json();
            setPublicDomain(data.domain);
        }
    } catch (error) {
        console.error("Failed to fetch public config", error);
    }
  };

  const fetchVehicles = async () => {
    const token = localStorage.getItem("token");
    if (!token) return;
    
    try {
        const response = await fetch(`${API_URL}/vehicles`, {
            headers: { Authorization: `Bearer ${token}` }
        });
        if (response.ok) {
            const data = await response.json();
            setVehicles(data);
        }
    } catch (error) {
        console.error("Failed to fetch vehicles", error);
    }
  };

  const fetchBrands = async () => {
    const token = localStorage.getItem("token");
    if (!token) return;

    try {
        const response = await fetch(`${API_URL}/brands`, {
            headers: { Authorization: `Bearer ${token}` }
        });
        if (response.ok) {
            const data = await response.json();
            setBrands(data);
        }
    } catch (error) {
        console.error("Failed to fetch brands", error);
    }
  };

  useEffect(() => {
    fetchVehicles();
    fetchBrands();
    fetchPublicConfig();
  }, []);

  const handleOpenAdd = () => {
    setEditingVehicle(null);
    const initialBrand = brands.length > 0 ? brands[0].name : "";
    setFormData({
        name: "",
        brand: initialBrand,
        locationType: LOCATIONS[0],
        customLocation: "",
        location: LOCATIONS[0],
        data: "{}",
        enabled: true
    });
    setApiRegion("");
    setDynamicFormData({});
    setTeslaTokenMode("manual");
    setTeslaVerifier("");
    setTeslaCallbackUrl("");
    setCurrentStep(0);
    setTeslaProducts([]);
    onOpen();
  };

  const handleOpenEdit = (vehicle: Vehicle) => {
    setEditingVehicle(vehicle);
    const isCustom = !LOCATIONS.includes(vehicle.location) && vehicle.location !== "";
    
    let parsedData: any = {};
    try {
        parsedData = JSON.parse(vehicle.data);
    } catch (e) {
        console.error("Failed to parse vehicle data", e);
    }

    setFormData({
        name: vehicle.name,
        brand: vehicle.brand,
        locationType: isCustom ? "Custom" : (vehicle.location || LOCATIONS[0]),
        customLocation: isCustom ? vehicle.location : "",
        location: vehicle.location,
        data: vehicle.data,
        enabled: vehicle.enabled
    });
    setApiRegion(parsedData.api || "");
    setDynamicFormData(parsedData);
    setTeslaTokenMode("manual");
    setTeslaVerifier("");
    setTeslaCallbackUrl("");
    setCurrentStep(0);
    setTeslaProducts([]);
    onOpen();
  };

  const handleOpenQR = (vehicle: Vehicle) => {
    if (!vehicle.qr_code?.uuid) return;
    
    // Prefer current window origin to support local network access
    const origin = typeof window !== "undefined" && window.location.origin && !window.location.origin.includes("localhost") 
        ? window.location.origin 
        : (publicDomain || "http://localhost:3000");
        
    const url = `${origin}/aed?uuid=${vehicle.qr_code.uuid}`;
    setQrCodeData(url);
    setSelectedVehicleName(vehicle.name);
    onQROpen();
  };

  const handleDelete = async (id: number) => {
    if (!confirm(t("dashboard.confirm_delete"))) return;
    
    const token = localStorage.getItem("token");
    try {
        const response = await fetch(`${API_URL}/vehicles/${id}`, {
            method: "DELETE",
            headers: { Authorization: `Bearer ${token}` }
        });
        if (response.ok) {
            fetchVehicles();
        }
    } catch (error) {
        console.error("Failed to delete vehicle", error);
    }
  };

  const handleSubmit = async (onClose: () => void) => {
    if (!apiRegion) {
        alert(t("dashboard.api_region_required", { defaultValue: "API Region is required" }));
        return;
    }
    const token = localStorage.getItem("token");
    const url = editingVehicle 
        ? `${API_URL}/vehicles/${editingVehicle.ID}` 
        : `${API_URL}/vehicles`;
    const method = editingVehicle ? "PUT" : "POST";

    // Determine final location value
    const finalLocation = formData.locationType === "Custom" ? formData.customLocation : formData.locationType;

    try {
        const response = await fetch(url, {
            method: method,
            headers: { 
                "Content-Type": "application/json",
                Authorization: `Bearer ${token}` 
            },
            body: JSON.stringify({
                ...formData,
                location: finalLocation,
                data: JSON.stringify({ ...dynamicFormData, api: apiRegion })
            })
        });

        if (response.ok) {
            fetchVehicles();
            onClose();
        } else {
            const err = await response.json();
            alert(err.error || "Operation failed");
        }
    } catch (error) {
        console.error("Failed to save vehicle", error);
    }
  };

  const handleTeslaLogin = async () => {
    const token = localStorage.getItem("token");
    if (!token) return;

    setTeslaTokenLoading(true);
    try {
        const response = await fetch(`${API_URL}/tesla/auth-url`, {
            headers: { Authorization: `Bearer ${token}` }
        });
        if (response.ok) {
            const data = await response.json();
            setTeslaVerifier(data.verifier);
            // Open in new tab
            window.open(data.url, "_blank");
        } else {
            alert("Failed to get Tesla Auth URL");
        }
    } catch (error) {
        console.error("Tesla auth error", error);
    } finally {
        setTeslaTokenLoading(false);
    }
  };

  const handleTeslaExchange = async () => {
    if (!teslaCallbackUrl) return;
    
    // Extract code
    let code = "";
    try {
        const url = new URL(teslaCallbackUrl);
        code = url.searchParams.get("code") || "";
    } catch (e) {
        // Try if the user pasted just the code? Or just invalid URL
        if (teslaCallbackUrl.startsWith("http")) {
             alert("Invalid URL");
             return;
        }
        // Maybe they pasted just the code? Let's assume URL first as per instruction
        code = teslaCallbackUrl;
    }

    if (!code) {
        alert("No code found in URL");
        return;
    }

    const token = localStorage.getItem("token");
    setTeslaTokenLoading(true);

    try {
        const response = await fetch(`${API_URL}/tesla/exchange`, {
            method: "POST",
            headers: { 
                "Content-Type": "application/json",
                Authorization: `Bearer ${token}` 
            },
            body: JSON.stringify({
                code,
                verifier: teslaVerifier
            })
        });

        if (response.ok) {
            const data = await response.json();
            // Fill the dynamic form
            setDynamicFormData(prev => ({
                ...prev,
                "access_token": data.access_token,
                "refresh_token": data.refresh_token,
                "api": apiRegion
            }));
            alert(t("dashboard.tesla_token_success", { defaultValue: "Tokens retrieved successfully!" }));
            setTeslaTokenMode("manual"); // Switch back to show the tokens
        } else {
            const err = await response.json();
            alert(err.error || "Failed to exchange token");
        }
    } catch (error) {
        console.error("Token exchange failed", error);
    } finally {
        setTeslaTokenLoading(false);
    }
  };

  const fetchTeslaVehicles = async () => {
    const token = localStorage.getItem("token");
    if (!token) return;

    setFetchVehicles(true);
    try {
        const response = await fetch(`${API_URL}/tesla/products`, {
            method: "POST",
            headers: { 
                "Content-Type": "application/json",
                Authorization: `Bearer ${token}` 
            },
            body: JSON.stringify({
                access_token: dynamicFormData.access_token,
                refresh_token: dynamicFormData.refresh_token,
                api_region: apiRegion
            })
        });

        if (response.ok) {
            const data = await response.json();
            setTeslaProducts(data.products || []);
            // Update tokens if refreshed
            if (data.access_token !== dynamicFormData.access_token) {
                setDynamicFormData(prev => ({
                    ...prev,
                    access_token: data.access_token,
                    refresh_token: data.refresh_token
                }));
            }
        } else {
            const err = await response.json();
            alert(err.error || "Failed to fetch Tesla vehicles");
        }
    } catch (error) {
        console.error("Failed to fetch vehicles", error);
    } finally {
        setFetchVehicles(false);
    }
  };

  const handleSelectTeslaProduct = (product: any) => {
    setDynamicFormData(prev => ({
        ...prev,
        vin: product.vin,
        id: String(product.id)
    }));
    // Auto set vehicle name if empty
    if (!formData.name) {
        setFormData(prev => ({ ...prev, name: product.display_name }));
    }
  };

  // Helper to get current brand config
  const currentBrand = brands.find(b => b.name === formData.brand);

  return (
    <MainLayout>
        <div className="flex justify-between items-center mb-6">
            <h1 className="text-2xl font-bold">{t("dashboard.title")}</h1>
        </div>
        
        <div className="grid gap-4 md:grid-cols-2 lg:grid-cols-3">
            {vehicles.map((vehicle) => (
                <div key={vehicle.ID} className="relative group">
                    <Card isPressable className="w-full cursor-pointer hover:bg-content2 transition-colors" onPress={() => handleOpenEdit(vehicle)}>
                        <CardBody>
                            <div className="flex justify-between items-start w-full">
                                <div className="flex flex-col items-start gap-1">
                                    <h3 className="text-lg font-bold">{vehicle.name}</h3>
                                    <div className="flex gap-2 text-sm text-default-500">
                                        <span>{vehicle.brand}</span>
                                        <span>•</span>
                                        <span>{new Date(vehicle.CreatedAt).toLocaleDateString()}</span>
                                    </div>
                                </div>
                                <div className="flex flex-col items-end gap-2">
                                    <div className={`w-3 h-3 rounded-full ${vehicle.enabled ? 'bg-success' : 'bg-danger'}`}></div>
                                </div>
                            </div>
                        </CardBody>
                    </Card>
                    <div className="absolute bottom-2 right-2 z-10">
                        <Button 
                            isIconOnly 
                            size="sm" 
                            color="primary" 
                            variant="flat"
                            onPress={() => handleOpenQR(vehicle)}
                        >
                            <FaQrcode />
                        </Button>
                    </div>
                </div>
            ))}
            {vehicles.length === 0 && (
                <div className="col-span-full text-center py-10 text-default-500">
                    {t("dashboard.no_vehicles")}
                </div>
            )}
        </div>

      {/* Floating Action Button */}
      <div className="fixed bottom-6 right-6 z-50">
        <Button 
            isIconOnly 
            color="primary" 
            size="lg" 
            className="rounded-full shadow-lg w-14 h-14"
            onPress={handleOpenAdd}
            aria-label={t("dashboard.add_vehicle")}
        >
            <FaPlus size={24} />
        </Button>
      </div>

      {/* Add/Edit Vehicle Modal */}
      <Modal isOpen={isOpen} onOpenChange={onOpenChange}>
        <ModalContent>
          {(onClose) => (
            <>
              <ModalHeader className="flex flex-col gap-1">
                {editingVehicle ? t("dashboard.edit_vehicle") : t("dashboard.add_vehicle")}
              </ModalHeader>
              <ModalBody>
                <div className="flex flex-col gap-4">
                    {/* Steps Navigation */}
                    {formData.brand === "Tesla" && (
                        <div className="flex justify-between items-center px-2 mb-2">
                            <div className={`flex items-center ${currentStep >= 0 ? "text-primary font-bold" : "text-default-400"}`}>
                                <span className="w-6 h-6 rounded-full border flex items-center justify-center mr-2 text-xs">1</span>
                                {t("dashboard.step_basic")}
                            </div>
                            <div className="h-[1px] bg-default-300 flex-1 mx-2"></div>
                            <div className={`flex items-center ${currentStep >= 1 ? "text-primary font-bold" : "text-default-400"}`}>
                                <span className="w-6 h-6 rounded-full border flex items-center justify-center mr-2 text-xs">2</span>
                                {t("dashboard.step_token")}
                            </div>
                            <div className="h-[1px] bg-default-300 flex-1 mx-2"></div>
                            <div className={`flex items-center ${currentStep >= 2 ? "text-primary font-bold" : "text-default-400"}`}>
                                <span className="w-6 h-6 rounded-full border flex items-center justify-center mr-2 text-xs">3</span>
                                {t("dashboard.step_vehicle")}
                            </div>
                        </div>
                    )}

                    {/* Step 0: Basic Info */}
                    {(formData.brand !== "Tesla" || currentStep === 0) && (
                        <>
                            <Input 
                                label={t("dashboard.vehicle_name")} 
                                value={formData.name}
                                onValueChange={(val) => setFormData({...formData, name: val})}
                                isRequired
                            />
                            <Select 
                                label={t("dashboard.vehicle_brand")} 
                                selectedKeys={formData.brand ? [formData.brand] : []}
                                onSelectionChange={(keys) => {
                                    const newBrand = Array.from(keys)[0] as string;
                                    setFormData({...formData, brand: newBrand});
                                    setDynamicFormData({});
                                }}
                            >
                                {brands.map((brand) => (
                                    <SelectItem key={brand.name}>{brand.name}</SelectItem>
                                ))}
                            </Select>
                            <Select 
                                label={t("dashboard.vehicle_location")} 
                                selectedKeys={formData.locationType ? [formData.locationType] : []}
                                onSelectionChange={(keys) => setFormData({...formData, locationType: Array.from(keys)[0] as string})}
                                isRequired
                            >
                                {LOCATIONS.map((loc) => (
                                    <SelectItem key={loc}>{t(`locations.${loc}`, { defaultValue: loc })}</SelectItem>
                                ))}
                            </Select>
                            {formData.locationType === "Custom" && (
                                <Input
                                    label={t("dashboard.custom_location")}
                                    value={formData.customLocation}
                                    onValueChange={(val) => setFormData({...formData, customLocation: val})}
                                    isRequired
                                />
                            )}
                            <div className="flex justify-between items-center px-1">
                                <span>{t("dashboard.enabled")}</span>
                                <Switch 
                                    isSelected={formData.enabled} 
                                    onValueChange={(val) => setFormData({...formData, enabled: val})}
                                />
                            </div>
                        </>
                    )}

                    {/* Step 1: Tesla Tokens */}
                    {formData.brand === "Tesla" && currentStep === 1 && (
                        <>
                            <div className="flex flex-col gap-2">
                                <span className="text-small text-default-500">{t("dashboard.api_region")} <span className="text-danger">*</span></span>
                                <div className="flex gap-2">
                                    <Button 
                                        color={apiRegion === "auth.tesla.cn" ? "primary" : "default"}
                                        variant={apiRegion === "auth.tesla.cn" ? "solid" : "flat"}
                                        onPress={() => setApiRegion("auth.tesla.cn")}
                                        className="flex-1"
                                    >
                                        {t("dashboard.api_china")}
                                    </Button>
                                    <Button 
                                        color={apiRegion === "auth.tesla.com" ? "primary" : "default"}
                                        variant={apiRegion === "auth.tesla.com" ? "solid" : "flat"}
                                        onPress={() => setApiRegion("auth.tesla.com")}
                                        className="flex-1"
                                    >
                                        {t("dashboard.api_intl")}
                                    </Button>
                                </div>
                            </div>
                            
                            <div className="flex flex-col gap-2 border p-3 rounded-lg border-default-200">
                                 <span className="text-small font-bold">{t("dashboard.tesla_token_mode")}<span className="text-danger">*</span></span>
                                 <div className="flex gap-2 mb-2">
                                    <Button 
                                        size="sm"
                                        color={teslaTokenMode === "manual" ? "primary" : "default"}
                                        variant={teslaTokenMode === "manual" ? "solid" : "flat"}
                                        onPress={() => setTeslaTokenMode("manual")}
                                        className="flex-1"
                                    >
                                        {t("dashboard.tesla_token_manual")}
                                    </Button>
                                    <Button 
                                        size="sm"
                                        color={teslaTokenMode === "auto" ? "primary" : "default"}
                                        variant={teslaTokenMode === "auto" ? "solid" : "flat"}
                                        onPress={() => setTeslaTokenMode("auto")}
                                        className="flex-1"
                                    >
                                        {t("dashboard.tesla_token_auto")}
                                    </Button>
                                 </div>

                                 {teslaTokenMode === "auto" && (
                                    <div className="flex flex-col gap-3">
                                        <Button 
                                            color="primary" 
                                            onPress={handleTeslaLogin} 
                                            isLoading={teslaTokenLoading}
                                        >
                                            {t("dashboard.tesla_login_btn")}
                                        </Button>
                                        
                                        <Input
                                            label={t("dashboard.tesla_paste_url")}
                                            placeholder="https://auth.tesla.com/void/callback?code=..."
                                            value={teslaCallbackUrl}
                                            onValueChange={(val) => {
                                                setTeslaCallbackUrl(val);
                                                // Auto-detect region
                                                let code = "";
                                                try {
                                                    const url = new URL(val);
                                                    code = url.searchParams.get("code") || "";
                                                } catch (e) {
                                                    if (!val.startsWith("http")) {
                                                         code = val;
                                                    }
                                                }
                                                if (code) {
                                                    if (code.startsWith("CN_")) {
                                                        setApiRegion("auth.tesla.cn");
                                                    } else {
                                                        setApiRegion("auth.tesla.com");
                                                    }
                                                }
                                            }}
                                            description="Copy the full URL from the 'Page Not Found' page after login"
                                        />
                                        
                                        <Button 
                                            color="success" 
                                            className="text-white"
                                            onPress={handleTeslaExchange}
                                            isLoading={teslaTokenLoading}
                                            isDisabled={!teslaCallbackUrl}
                                        >
                                            {t("dashboard.tesla_exchange_btn")}
                                        </Button>
                                    </div>
                                 )}
                            </div>

                            {/* Manual inputs always visible or only in manual mode? User requirement says 2 buttons manual/auto. If manual selected, show inputs. If auto selected, show flow. But inputs must be populated eventually. */}
                            <Input
                                label="Access Token"
                                value={dynamicFormData["access_token"] || ""}
                                onValueChange={(val) => setDynamicFormData(prev => ({ ...prev, "access_token": val }))}
                                isReadOnly={teslaTokenMode === "auto"}
                            />
                            <Input
                                label="Refresh Token"
                                value={dynamicFormData["refresh_token"] || ""}
                                onValueChange={(val) => setDynamicFormData(prev => ({ ...prev, "refresh_token": val }))}
                                isReadOnly={teslaTokenMode === "auto"}
                            />
                        </>
                    )}

                    {/* Step 2: Vehicle Selection */}
                    {formData.brand === "Tesla" && currentStep === 2 && (
                        <div className="flex flex-col gap-4">
                            <Button 
                                color="secondary" 
                                onPress={fetchTeslaVehicles} 
                                isLoading={fetchingVehicles}
                            >
                                {t("dashboard.fetch_vehicles")}
                            </Button>
                            
                            <div className="max-h-[300px] overflow-y-auto border rounded-lg p-2">
                                {teslaProducts.length === 0 ? (
                                    <div className="text-center text-default-500 py-4">
                                        No vehicles found. Click Fetch Vehicles.
                                    </div>
                                ) : (
                                    teslaProducts.map((p) => (
                                        <div 
                                            key={p.id} 
                                            className={`p-3 border-b last:border-b-0 cursor-pointer hover:bg-default-100 flex justify-between items-center ${String(p.id) === dynamicFormData.id ? "bg-primary-50" : ""}`}
                                            onClick={() => handleSelectTeslaProduct(p)}
                                        >
                                            <div>
                                                <div className="font-bold">{p.display_name}</div>
                                                <div className="text-xs text-default-500">VIN: {p.vin}</div>
                                            </div>
                                            {String(p.id) === dynamicFormData.id && (
                                                <div className="text-primary text-sm font-bold">Selected</div>
                                            )}
                                        </div>
                                    ))
                                )}
                            </div>
                            
                            <Input
                                label="VIN"
                                value={dynamicFormData["vin"] || ""}
                                isReadOnly
                            />
                            <Input
                                label="Vehicle ID"
                                value={dynamicFormData["id"] || ""}
                                isReadOnly
                            />
                        </div>
                    )}
                    
                    {/* Other brands dynamic fields */}
                    {formData.brand !== "Tesla" && currentBrand && currentBrand.fields && Object.entries(currentBrand.fields).map(([label, jsonKey]) => (
                        <Input
                            key={jsonKey}
                            label={label}
                            value={dynamicFormData[jsonKey] || ""}
                            onValueChange={(val) => setDynamicFormData(prev => ({
                                ...prev,
                                [jsonKey]: val
                            }))}
                        />
                    ))}
                </div>
              </ModalBody>
              <ModalFooter>
                {/* Delete button only on step 0 or if editing */}
                {editingVehicle && (
                    <Button color="danger" variant="light" onPress={() => {
                        handleDelete(editingVehicle.ID);
                        onClose();
                    }}>
                        <FaTrash className="mr-2" />
                        {t("dashboard.delete")}
                    </Button>
                )}
                <div className="flex-grow"></div>
                
                <Button variant="light" onPress={onClose}>
                  {t("common.cancel")}
                </Button>

                {formData.brand === "Tesla" ? (
                    <>
                        {currentStep > 0 && (
                            <Button onPress={() => setCurrentStep(currentStep - 1)}>
                                {t("dashboard.prev_step")}
                            </Button>
                        )}
                        {currentStep < 2 ? (
                            <Button 
                                color="primary" 
                                onPress={() => {
                                    // Validation before next
                                    if (currentStep === 1 && (!dynamicFormData.access_token || !apiRegion)) {
                                        alert("Please provide tokens and region");
                                        return;
                                    }
                                    setCurrentStep(currentStep + 1);
                                }}
                            >
                                {t("dashboard.next_step")}
                            </Button>
                        ) : (
                            <Button color="primary" onPress={() => handleSubmit(onClose)}>
                                {t("common.save")}
                            </Button>
                        )}
                    </>
                ) : (
                    <Button color="primary" onPress={() => handleSubmit(onClose)}>
                      {t("common.save")}
                    </Button>
                )}
              </ModalFooter>
            </>
          )}
        </ModalContent>
      </Modal>

      {/* QR Code Modal */}
      <Modal isOpen={isQRModalOpen} onOpenChange={onQROpenChange}>
        <ModalContent>
          {(onClose) => (
            <>
              <ModalHeader className="flex flex-col gap-1">
                {t("dashboard.qr_code")} - {selectedVehicleName}
              </ModalHeader>
              <ModalBody>
                <div className="flex justify-center p-4 bg-white rounded-lg">
                    {qrCodeData && (
                        <QRCode value={qrCodeData} size={256} />
                    )}
                </div>
                <p className="text-center text-sm text-default-500 break-all">{qrCodeData}</p>
              </ModalBody>
              <ModalFooter>
                <Button color="primary" onPress={onClose}>
                  {t("common.close")}
                </Button>
              </ModalFooter>
            </>
          )}
        </ModalContent>
      </Modal>
    </MainLayout>
  );
}
