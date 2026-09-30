import { useEffect, useRef, useState } from "react"
import { boundsFromMap, type LngLatBounds, type YandexMap, type YandexPlacemark, type YMaps } from "./ymaps"

export function useYandexMap(onSelect: (id: string) => void) {
  const containerRef = useRef<HTMLDivElement>(null)
  const mapRef = useRef<YandexMap | null>(null)
  const markersRef = useRef<Map<string, YandexPlacemark>>(new Map())
  const zoomRef = useRef(13)
  const onSelectRef = useRef(onSelect)
  const [bounds, setBounds] = useState<LngLatBounds | null>(null)
  const [ready, setReady] = useState(false)
  const [error, setError] = useState<string | null>(null)

  useEffect(() => {
    onSelectRef.current = onSelect
  }, [onSelect])

  useEffect(() => {
    const element = containerRef.current
    if (!element) return

    let map: YandexMap | null = null
    let cancelled = false
    let destroyed = false
    let frame = 0
    let pending: LngLatBounds | null = null
    let placeZoom: (() => void) | null = null

    const destroy = () => {
      if (destroyed) return
      destroyed = true
      cancelAnimationFrame(frame)
      if (placeZoom) window.removeEventListener("resize", placeZoom)
      map?.destroy()
      map = null
      mapRef.current = null
      markersRef.current.clear()
    }

    waitForYmaps()
      .then(
        (ymaps) =>
          new Promise<YMaps>((resolve) => {
            ymaps.ready(() => resolve(ymaps))
          }),
      )
      .then((ymaps) => {
        if (cancelled) return
        map = new ymaps.Map(
          element,
          { center: [55.7558, 37.6173], zoom: 13, controls: [] },
          { suppressMapOpenBlock: true },
        )
        const zoom = map.controls.add("zoomControl", { size: "small", position: zoomPosition() })
        placeZoom = () => zoom.options.set("position", zoomPosition())
        window.addEventListener("resize", placeZoom)
        map.events.add("boundschange", () => {
          if (cancelled || !map) return
          pending = boundsFromMap(map.getBounds())
          zoomRef.current = map.getZoom()
          if (frame) return
          frame = requestAnimationFrame(() => {
            frame = 0
            if (!cancelled && pending) setBounds(pending)
          })
        })
        if (cancelled) {
          destroy()
          return
        }
        mapRef.current = map
        zoomRef.current = map.getZoom()
        setBounds(boundsFromMap(map.getBounds()))
        setReady(true)
      })
      .catch(() => {
        if (!cancelled) setError("JavaScript API Яндекс Карт не загрузился")
      })

    return () => {
      cancelled = true
      destroy()
    }
  }, [])

  return { containerRef, mapRef, markersRef, zoomRef, onSelectRef, bounds, ready, error }
}

function zoomPosition(): { left: number; top: number } {
  if (window.innerWidth <= 720) return { left: 16, top: 16 }
  return { left: 392, top: 16 }
}

function waitForYmaps(): Promise<YMaps> {
  if (window.ymaps) return Promise.resolve(window.ymaps)
  return new Promise((resolve, reject) => {
    const started = Date.now()
    const timer = window.setInterval(() => {
      if (window.ymaps) {
        window.clearInterval(timer)
        resolve(window.ymaps)
      } else if (Date.now() - started > 8000) {
        window.clearInterval(timer)
        reject(new Error("ymaps timeout"))
      }
    }, 50)
  })
}
