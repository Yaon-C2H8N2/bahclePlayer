import { useEffect, useState } from "react";
import { fetchApi } from "@/lib/network.ts";
import { LoaderCircle, Copy, Check, Save } from "lucide-react";
import { Button } from "@/components/ui/button.tsx";
import { Input } from "@/components/ui/input.tsx";
import { toast } from "@/hooks/use-toast.ts";

interface IOverlaysProps {
    onClose: () => void;
}

interface OverlayType {
    overlay_type_id: number;
    name: string;
    description: string;
    schema: any;
    overlay_code: string;
    link: string;
}

interface UserOverlaySettings {
    css?: string;
}

export const Overlays = (props: IOverlaysProps) => {
    const [loading, setLoading] = useState<boolean>(true);
    const [overlayTypes, setOverlayTypes] = useState<OverlayType[]>([]);
    const [selectedOverlayType, setSelectedOverlayType] = useState<OverlayType | null>(null);
    const [copied, setCopied] = useState<boolean>(false);
    const [customCss, setCustomCss] = useState<string>("");
    const [savingSettings, setSavingSettings] = useState<boolean>(false);
    const [_userSettings, setUserSettings] = useState<UserOverlaySettings>({});

    useEffect(() => {
        if (!document.cookie.includes("token")) {
            window.location.href = "/";
        }

        fetchApi("/api/overlays")
            .then((res) => res.json())
            .then((data) => {
                setOverlayTypes(data.overlay_types || []);
                setLoading(false);
            })
            .catch((error) => {
                console.error("Error fetching overlay types:", error);
                setLoading(false);
            });
    }, []);

    const loadUserSettings = async (overlayCode: string) => {
        try {
            const response = await fetchApi(`/api/overlays/settings?overlay_code=${overlayCode}`);
            if (response.ok) {
                const data = await response.json();
                const settings = JSON.parse(data.settings || '{}');
                setUserSettings(settings);
                setCustomCss(settings.css || '');
            }
        } catch (error) {
            console.error("Failed to load user settings:", error);
        }
    };

    const handleOverlaySelect = (overlayType: OverlayType) => {
        setSelectedOverlayType(overlayType);
        if (overlayType.overlay_code === 'currently_playing') {
            loadUserSettings(overlayType.overlay_code);
        }
    };

    const handleCopyLink = async () => {
        if (selectedOverlayType) {
            try {
                await navigator.clipboard.writeText(selectedOverlayType.link);
                setCopied(true);
                setTimeout(() => setCopied(false), 2000);
            } catch (err) {
                console.error("Failed to copy link:", err);
            }
        }
    };

    const handleSaveSettings = async () => {
        if (!selectedOverlayType) return;

        setSavingSettings(true);
        try {
            const settings = {
                css: customCss
            };

            const response = await fetchApi("/api/overlays/settings", {
                method: "POST",
                headers: {
                    "Content-Type": "application/json",
                },
                body: JSON.stringify({
                    overlay_code: selectedOverlayType.overlay_code,
                    settings: settings
                }),
            });

            if (response.ok) {
                toast({
                    title: "Settings saved",
                    description: "Your overlay settings have been saved successfully.",
                });
                setUserSettings(settings);
            } else {
                throw new Error("Failed to save settings");
            }
        } catch (error) {
            console.error("Error saving settings:", error);
            toast({
                title: "Error",
                description: "Failed to save overlay settings. Please try again.",
                variant: "destructive",
            });
        } finally {
            setSavingSettings(false);
        }
    };

    return (
        <div className="flex flex-col h-full">
            {loading ? (
                <div className="flex justify-center items-center h-64">
                    <LoaderCircle className="animate-spin" />
                </div>
            ) : (
                <div className="flex flex-row h-[500px]">
                    {/* Left panel - Vertical list of overlay types */}
                    <div className="w-1/3 border-r pr-4 overflow-y-auto">
                        <h3 className="font-medium mb-2">Available Overlays</h3>
                        <div className="flex flex-col gap-2">
                            {overlayTypes.length === 0 ? (
                                <div className="text-sm text-gray-500">No overlay types available</div>
                            ) : (
                                overlayTypes.map((overlayType) => (
                                    <div
                                        key={overlayType.overlay_type_id}
                                        className={`p-2 rounded cursor-pointer hover:bg-gray-100 ${
                                            selectedOverlayType?.overlay_type_id === overlayType.overlay_type_id
                                                ? "bg-gray-100"
                                                : ""
                                        }`}
                                        onClick={() => handleOverlaySelect(overlayType)}
                                    >
                                        <div className="font-medium">{overlayType.name}</div>
                                        {overlayType.description && (
                                            <div className="text-sm text-gray-500">{overlayType.description}</div>
                                        )}
                                    </div>
                                ))
                            )}
                        </div>
                    </div>

                    {/* Right panel - Configuration */}
                    <div className="w-2/3 pl-4">
                        {selectedOverlayType ? (
                            <div className="space-y-4">
                                <div>
                                    <h3 className="font-medium mb-2">{selectedOverlayType.name}</h3>
                                    <div className="text-sm text-gray-500 mb-4">
                                        {selectedOverlayType.description || "No description available"}
                                    </div>
                                </div>

                                <div>
                                    <h4 className="font-medium mb-2">Overlay Link</h4>
                                    <div className="flex gap-2">
                                        <Input 
                                            value={selectedOverlayType.link} 
                                            readOnly 
                                            className="flex-grow"
                                        />
                                        <Button 
                                            onClick={handleCopyLink} 
                                            variant="outline" 
                                            className="flex gap-1 items-center"
                                        >
                                            {copied ? (
                                                <>
                                                    <Check className="h-4 w-4" />
                                                    Copied
                                                </>
                                            ) : (
                                                <>
                                                    <Copy className="h-4 w-4" />
                                                    Copy
                                                </>
                                            )}
                                        </Button>
                                    </div>
                                    <p className="text-sm text-gray-500 mt-1">
                                        Use this link in your streaming software as a browser source.
                                    </p>
                                </div>

                                {selectedOverlayType.overlay_code === 'currently_playing' && (
                                    <div>
                                        <h4 className="font-medium mb-2">Custom CSS</h4>
                                        <textarea
                                            value={customCss}
                                            onChange={(e) => setCustomCss(e.target.value)}
                                            className="w-full h-48 p-3 border rounded-md font-mono text-sm resize-none"
                                            placeholder="Enter your custom CSS here...

Example:
.video-card {
    background: linear-gradient(45deg, #ff6b6b, #4ecdc4);
    border-radius: 10px;
    padding: 20px;
}

.video-title {
    color: white;
    font-weight: bold;
}"
                                        />
                                        <div className="flex justify-between items-center mt-2">
                                            <p className="text-sm text-gray-500">
                                                Customize the appearance of your overlay with CSS
                                            </p>
                                            <Button
                                                onClick={handleSaveSettings}
                                                disabled={savingSettings}
                                                className="flex gap-1 items-center"
                                            >
                                                {savingSettings ? (
                                                    <>
                                                        <LoaderCircle className="h-4 w-4 animate-spin" />
                                                        Saving...
                                                    </>
                                                ) : (
                                                    <>
                                                        <Save className="h-4 w-4" />
                                                        Save Settings
                                                    </>
                                                )}
                                            </Button>
                                        </div>
                                    </div>
                                )}

                                {selectedOverlayType.overlay_code !== 'currently_playing' && (
                                    <div className="p-4 border rounded bg-gray-50 text-center text-gray-500">
                                        Configuration options will be available soon for this overlay
                                    </div>
                                )}
                            </div>
                        ) : (
                            <div className="flex justify-center items-center h-full text-gray-500">
                                Select an overlay type from the list
                            </div>
                        )}
                    </div>
                </div>
            )}

            <div className="flex flex-row justify-end gap-2 w-full mt-4">
                <Button variant="ghost" onClick={() => props.onClose()}>
                    Close
                </Button>
            </div>
        </div>
    );
};
