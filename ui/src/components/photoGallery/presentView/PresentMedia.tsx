import React, { useRef, useState, useEffect } from 'react'
import styled from 'styled-components'
import { MediaType } from '../../../__generated__/globalTypes'
import { exhaustiveCheck } from '../../../helpers/utils'
import { ProtectedImage, ProtectedVideo, getProtectedUrl } from '../ProtectedMedia'
import { MediaGalleryFields } from '../__generated__/MediaGalleryFields'

// Sized to the presentation frame rather than the viewport: the frame is the
// full viewport until the info panel is pinned beside it, and then the photo
// has to fit the space that is left instead of running underneath the panel.
const StyledPhoto = styled(ProtectedImage)`
  position: absolute;
  top: 0;
  left: 0;
  width: 100%;
  height: 100%;
  object-fit: contain;
  object-position: center;
`

const StyledVideo = styled(ProtectedVideo)`
  position: absolute;
  top: 0;
  left: 0;
  width: 100%;
  height: 100%;
`

const StyledLiveVideo = styled.video<{ $visible: boolean }>`
  position: absolute;
  top: 0;
  left: 0;
  width: 100%;
  height: 100%;
  object-fit: contain;
  object-position: center;
  pointer-events: none;
  opacity: ${({ $visible }) => ($visible ? 1 : 0)};
  transition: opacity 120ms ease-in-out;
  z-index: 2;
`

const LivePillButton = styled.button<{ $playing: boolean }>`
  position: absolute;
  top: 28px;
  left: 96px;
  z-index: 10;
  display: inline-flex;
  align-items: center;
  gap: 6px;
  padding: 6px 12px;
  border-radius: 9999px;
  background-color: ${({ $playing }) =>
    $playing ? 'rgba(255, 255, 255, 0.95)' : 'rgba(0, 0, 0, 0.5)'};
  color: ${({ $playing }) => ($playing ? '#000' : '#fff')};
  font-size: 12px;
  font-weight: 700;
  letter-spacing: 0.5px;
  border: 1px solid
    ${({ $playing }) =>
      $playing ? 'rgba(255, 255, 255, 1)' : 'rgba(255, 255, 255, 0.25)'};
  cursor: pointer;
  backdrop-filter: blur(8px);
  box-shadow: 0 2px 8px rgba(0, 0, 0, 0.4);
  user-select: none;
  transition: all 150ms ease-in-out;

  &:hover {
    background-color: ${({ $playing }) =>
      $playing ? 'rgba(255, 255, 255, 1)' : 'rgba(0, 0, 0, 0.75)'};
    transform: scale(1.04);
  }

  &:active {
    transform: scale(0.96);
  }
`

const LivePhotoIrisSVG = () => (
  <svg
    width="14"
    height="14"
    viewBox="0 0 24 24"
    fill="none"
    stroke="currentColor"
    strokeWidth="2.5"
    strokeLinecap="round"
    strokeLinejoin="round"
  >
    <circle cx="12" cy="12" r="9" strokeDasharray="3 3" />
    <circle cx="12" cy="12" r="5" />
    <circle cx="12" cy="12" r="1.5" fill="currentColor" />
  </svg>
)

type PresentLivePhotoProps = {
  media: MediaGalleryFields
  imageLoaded?(): void
}

const PresentLivePhoto = ({
  media,
  imageLoaded,
  ...otherProps
}: PresentLivePhotoProps) => {
  const [isPlaying, setIsPlaying] = useState(false)
  const videoRef = useRef<HTMLVideoElement>(null)
  const isPressing = useRef(false)

  const videoUrl = getProtectedUrl(media.livePhotoVideo?.url)

  useEffect(() => {
    setIsPlaying(false)
    if (videoRef.current) {
      videoRef.current.pause()
      videoRef.current.currentTime = 0
    }
  }, [media.id])

  const startPlayback = () => {
    if (!videoRef.current) return
    setIsPlaying(true)
    videoRef.current.currentTime = 0
    videoRef.current.play().catch(() => {
      if (videoRef.current) {
        videoRef.current.muted = true
        videoRef.current.play().catch(() => {})
      }
    })
  }

  const stopPlayback = () => {
    setIsPlaying(false)
    if (videoRef.current) {
      videoRef.current.pause()
      videoRef.current.currentTime = 0
    }
  }

  const handleToggleClick = (e: React.MouseEvent) => {
    e.stopPropagation()
    if (isPlaying) {
      stopPlayback()
    } else {
      startPlayback()
    }
  }

  const handleMouseDown = () => {
    isPressing.current = true
    startPlayback()
  }

  const handleMouseUp = () => {
    if (isPressing.current) {
      isPressing.current = false
      stopPlayback()
    }
  }

  return (
    <div
      {...otherProps}
      onMouseDown={handleMouseDown}
      onMouseUp={handleMouseUp}
      onMouseLeave={handleMouseUp}
      onTouchStart={handleMouseDown}
      onTouchEnd={handleMouseUp}
    >
      <LivePillButton
        $playing={isPlaying}
        onClick={handleToggleClick}
        onMouseDown={e => e.stopPropagation()}
        onTouchStart={e => e.stopPropagation()}
        title={isPlaying ? 'Pause Live Photo' : 'Play Live Photo'}
      >
        <LivePhotoIrisSVG />
        <span>LIVE</span>
      </LivePillButton>

      <StyledPhoto
        key={`${media.id}-thumb`}
        src={media.thumbnail?.url}
        data-testid="present-img-thumbnail"
      />
      <StyledPhoto
        key={`${media.id}-highres`}
        style={{ display: 'none' }}
        src={media.highRes?.url}
        data-testid="present-img-highres"
        onLoad={e => {
          const elem = e.target as HTMLImageElement
          elem.style.display = 'initial'
          imageLoaded && imageLoaded()
        }}
      />

      {videoUrl && (
        <StyledLiveVideo
          ref={videoRef}
          src={videoUrl}
          playsInline
          preload="metadata"
          $visible={isPlaying}
          onEnded={() => {
            isPressing.current = false
            stopPlayback()
          }}
          data-testid="present-live-video"
        />
      )}
    </div>
  )
}

type PresentMediaProps = {
  media: MediaGalleryFields
  imageLoaded?(): void
}

const PresentMedia = ({
  media,
  imageLoaded,
  ...otherProps
}: PresentMediaProps) => {
  switch (media.type) {
    case MediaType.Photo:
      if (media.livePhotoVideo) {
        return (
          <PresentLivePhoto
            media={media}
            imageLoaded={imageLoaded}
            {...otherProps}
          />
        )
      }

      return (
        <div {...otherProps}>
          <StyledPhoto
            key={`${media.id}-thumb`}
            src={media.thumbnail?.url}
            data-testid="present-img-thumbnail"
          />
          <StyledPhoto
            key={`${media.id}-highres`}
            style={{ display: 'none' }}
            src={media.highRes?.url}
            data-testid="present-img-highres"
            onLoad={e => {
              const elem = e.target as HTMLImageElement
              elem.style.display = 'initial'
              imageLoaded && imageLoaded()
            }}
          />
        </div>
      )
    case MediaType.Video:
      return <StyledVideo media={media} data-testid="present-video" />
  }

  exhaustiveCheck(media.type)
}

export default PresentMedia
