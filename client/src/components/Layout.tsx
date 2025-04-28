import {Outlet} from "react-router-dom";
import {Button} from "@/components/ui/button.tsx";
import {fetchApi} from "@/lib/network.ts";
import {SettingsDialog} from "@/pages/settings/SettingsDialog.tsx";
import {OverlaysDialog} from "@/pages/overlays/OverlaysDialog.tsx";
import {useState} from "react";

export const Layout = () => {
    const [openSettings, setOpenSettings] = useState<boolean>(false)
    const [openOverlays, setOpenOverlays] = useState<boolean>(false)

    const handleLogout = () => {
        fetchApi("/api/logout").then(() => {
            window.location.href = "/"
        })
    }

    return (
        <div className={"flex flex-col items-center w-full h-[100vh]"}>
            <div className={"flex flex-row w-full p-3 items-center justify-end gap-2"}>
                <Button variant={"ghost"} onClick={() => {
                    setOpenOverlays(true)
                }}>
                    Overlays
                </Button>
                <Button variant={"ghost"} onClick={() => {
                    setOpenSettings(true)
                }}>
                    Settings
                </Button>
                <Button onClick={() => {
                    handleLogout()
                }}>Log out</Button>
            </div>
            <OverlaysDialog open={openOverlays} onClose={() => setOpenOverlays(false)}/>
            <SettingsDialog open={openSettings} onClose={() => setOpenSettings(false)}/>
            <Outlet/>
        </div>
    )
}
