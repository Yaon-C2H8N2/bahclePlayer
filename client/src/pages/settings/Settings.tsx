import {useEffect, useState} from "react";
import {Checkbox} from "@/components/ui/checkbox.tsx";
import {Button} from "@/components/ui/button.tsx";
import {LoaderCircle} from "lucide-react";
import {fetchApi} from "@/lib/network.ts";
import {Select, SelectValue, SelectTrigger, SelectContent, SelectItem} from "@/components/ui/select.tsx";

interface ISettingsProps {
    onClose: () => void
}

export const Settings = (props: ISettingsProps) => {
    const [loading, setLoading] = useState<boolean>(true)
    const [rewards, setRewards] = useState<any>([])
    const [queueReward, setQueueReward] = useState<any>(null)
    const [playlistReward, setPlaylistReward] = useState<any>(null)
    const [queueMethod, setQueueMethod] = useState("")
    const [playlistMethod, setPlaylistMethod] = useState("")

    useEffect(() => {
        if (!document.cookie.includes("token")) {
            window.location.href = "/"
        }

        fetchApi("/api/rewards").then((res) => {
            return res.json()
        }).then((data) => {
            setRewards(data.rewards ?? [])

            const queueRewardSetting = data.settings?.find((setting: any) => setting.Config === "QUEUE_REDEMPTION")
            const playlistRewardSetting = data.settings?.find((setting: any) => setting.Config === "PLAYLIST_REDEMPTION")
            const queueMethodSetting = data.settings?.find((setting: any) => setting.Config === "QUEUE_METHOD")
            const playlistMethodSetting = data.settings?.find((setting: any) => setting.Config === "PLAYLIST_METHOD")

            setQueueReward(data.rewards?.find((reward: any) => reward.id === queueRewardSetting?.Value))
            setPlaylistReward(data.rewards?.find((reward: any) => reward.id === playlistRewardSetting?.Value))
            setQueueMethod(queueMethodSetting?.Value)
            setPlaylistMethod(playlistMethodSetting?.Value)

            setLoading(false)
        })
    }, [])

    const handleSave = () => {
        setLoading(true)
        fetchApi(`/api/settings?playlist_redemption=${playlistReward?.id ?? ""}&queue_redemption=${queueReward?.id ?? ""}&playlist_method=${playlistMethod ?? ""}&queue_method=${queueMethod ?? ""}`)
            .then(() => {
                setLoading(false)
                props.onClose()
            })
    }

    return (
        <div className={"flex flex-col justify-center items-center gap-4"}>
            {loading ? (
                <LoaderCircle className={"animate-spin"}/>
            ) : (
                <div className={"grid grid-cols-3 w-full gap-3"}>
                    <div>Reward</div>
                    <div>Queue</div>
                    <div>Playlist</div>
                    {rewards.map((reward: any) => {
                        return (
                            <>
                                <div>{reward.title}</div>
                                <div className={"flex flex-row items-center gap-2"}>
                                    <Checkbox
                                        checked={queueReward === reward}
                                        onCheckedChange={() => setQueueReward((prev: any) => {
                                            if (prev === reward) {
                                                return null
                                            }
                                            if (playlistReward === reward) {
                                                setPlaylistReward(null)
                                            }
                                            return reward
                                        })}
                                    />
                                    {queueReward === reward && (
                                        <Select value={queueMethod} onValueChange={(value) => setQueueMethod(value)}>
                                            <SelectTrigger>
                                                <SelectValue placeholder="Mode" />
                                            </SelectTrigger>
                                            <SelectContent>
                                                <SelectItem value="DIRECT">Direct</SelectItem>
                                                <SelectItem value="POLL">Poll</SelectItem>
                                            </SelectContent>
                                        </Select>
                                    )}
                                </div>
                                <div className={"flex flex-row items-center gap-2"}>
                                    <Checkbox
                                        checked={playlistReward === reward}
                                        onCheckedChange={() => setPlaylistReward((prev: any) => {
                                            if (prev === reward) {
                                                return null
                                            }
                                            if (queueReward === reward) {
                                                setQueueReward(null)
                                            }
                                            return reward
                                        })}
                                    />
                                    {playlistReward === reward && (
                                        <Select value={playlistMethod} onValueChange={(value) => setPlaylistMethod(value)}>
                                            <SelectTrigger>
                                                <SelectValue placeholder="Mode" />
                                            </SelectTrigger>
                                            <SelectContent>
                                                <SelectItem value="DIRECT">Direct</SelectItem>
                                                <SelectItem value="POLL">Poll</SelectItem>
                                            </SelectContent>
                                        </Select>
                                    )}
                                </div>
                            </>
                        )
                    })}
                    {rewards.length === 0 && <div className={"col-span-3 flex flex-row justify-center"}>No channel points rewards found</div>}
                </div>
            )}
            <div className={"flex flex-row justify-end gap-2 w-full"}>
                <Button variant={"ghost"} onClick={() => props.onClose()}>Cancel</Button>
                <Button disabled={loading} onClick={() => handleSave()}>Save</Button>
            </div>
        </div>
    )
}