import {useEffect, useState} from "react";
import {Checkbox} from "@/components/ui/checkbox.tsx";
import {Button} from "@/components/ui/button.tsx";
import {LoaderCircle} from "lucide-react";
import {fetchApi} from "@/lib/network.ts";

interface ISettingsProps {
    onClose: () => void
}

export const Settings = (props: ISettingsProps) => {
    const [loading, setLoading] = useState<boolean>(true)
    const [rewards, setRewards] = useState<any>([])
    const [queueReward, setQueueReward] = useState<any>(null)
    const [playlistReward, setPlaylistReward] = useState<any>(null)

    useEffect(() => {
        if (!document.cookie.includes("token")) {
            window.location.href = "/"
        }

        fetchApi("/api/rewards").then((res) => {
            return res.json()
        }).then((data) => {
            setRewards(data)
            setLoading(false)
        })
    }, [])

    const handleSave = () => {
        setLoading(true)
        fetchApi(`/api/settings?playlist_redemption=${playlistReward.id}&queue_redemption=${queueReward.id}`).then(() => {
            setLoading(false)
            props.onClose()
        })
    }

    return (
        <div className={"flex flex-col justify-center items-center gap-4"}>
            {loading ? (
                <LoaderCircle className={"animate-spin"}/>
            ) : (
                <div className={"grid grid-cols-3 w-full gap-2"}>
                    <div>Reward</div>
                    <div>Queue</div>
                    <div>Playlist</div>
                    {rewards.map((reward: any) => {
                        return (
                            <>
                                <div>{reward.title}</div>
                                <div>
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
                                </div>
                                <div>
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
                                </div>
                            </>
                        )
                    })}
                </div>
            )}
            <div className={"flex flex-row justify-end gap-2 w-full"}>
                <Button variant={"ghost"} onClick={() => props.onClose()}>Cancel</Button>
                <Button disabled={loading} onClick={() => handleSave()}>Save</Button>
            </div>
        </div>
    )
}