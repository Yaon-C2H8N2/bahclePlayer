import { useEffect, useState } from "react";
import { fetchApi } from "@/lib/network.ts";
import { LoaderCircle } from "lucide-react";
import { Button } from "@/components/ui/button.tsx";

interface IOverlaysProps {
    onClose: () => void;
}

interface OverlayType {
    overlay_type_id: number;
    name: string;
    description: string;
    schema: any;
}

export const Overlays = (props: IOverlaysProps) => {
    const [loading, setLoading] = useState<boolean>(true);
    const [overlayTypes, setOverlayTypes] = useState<OverlayType[]>([]);
    const [selectedOverlayType, setSelectedOverlayType] = useState<OverlayType | null>(null);

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

    return (
        <div className="flex flex-col h-full">
            {loading ? (
                <div className="flex justify-center items-center h-64">
                    <LoaderCircle className="animate-spin" />
                </div>
            ) : (
                <div className="flex flex-row h-[400px]">
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
                                        onClick={() => setSelectedOverlayType(overlayType)}
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

                    {/* Right panel - Blank for now */}
                    <div className="w-2/3 pl-4">
                        {selectedOverlayType ? (
                            <div>
                                <h3 className="font-medium mb-2">{selectedOverlayType.name}</h3>
                                <div className="text-sm text-gray-500 mb-4">
                                    {selectedOverlayType.description || "No description available"}
                                </div>
                                {/* Content will be added in future tasks */}
                                <div className="p-4 border rounded bg-gray-50 text-center text-gray-500">
                                    Overlay configuration will be available soon
                                </div>
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