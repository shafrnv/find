export type LngLat = [lng: number, lat: number]
export type LngLatBounds = [southWest: LngLat, northEast: LngLat]

export interface YandexPlacemark {
  events: { add(name: string, handler: () => void): void }
  options: { set(key: string, value: string): void }
}

export interface YandexMap {
  geoObjects: {
    add(object: YandexPlacemark): void
    remove(object: YandexPlacemark): void
  }
  events: { add(name: string, handler: () => void): void }
  controls: { add(name: string, options?: object): { options: { set(key: string, value: object): void } } }
  getBounds(): number[][]
  getZoom(): number
  setCenter(center: [number, number], zoom?: number, options?: { duration?: number }): void
  destroy(): void
}

export interface YMaps {
  ready(handler: () => void): void
  Map: new (
    element: HTMLElement,
    state: { center: [number, number]; zoom: number; controls: string[] },
    options?: { suppressMapOpenBlock?: boolean },
  ) => YandexMap
  Placemark: new (
    geometry: [number, number],
    properties?: { hintContent?: string },
    options?: { preset?: string },
  ) => YandexPlacemark
}

declare global {
  interface Window {
    ymaps?: YMaps
  }
}

export function toMapPoint(coordinates: LngLat): [lat: number, lng: number] {
  return [coordinates[1], coordinates[0]]
}

export function boundsFromMap(raw: number[][]): LngLatBounds {
  const [[latA, lngA], [latB, lngB]] = raw
  return [
    [Math.min(lngA, lngB), Math.min(latA, latB)],
    [Math.max(lngA, lngB), Math.max(latA, latB)],
  ]
}
