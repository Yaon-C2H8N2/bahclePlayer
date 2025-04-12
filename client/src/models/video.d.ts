interface Video {
    video_id: number
    user_id: number
    youtube_id: string
    url: string
    title: string
    duration: string
    type: string | "QUEUE" | "PLAYLIST"
    created_at: Date
    thumbnail_url: string
    added_by: string
}