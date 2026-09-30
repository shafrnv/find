import type { Intent, SubjectKind } from "../domain"
import type { LngLat, LngLatBounds } from "./ymaps"

export interface TraitLine {
  name: string
  value: string
}

export interface MapEntry {
  id: string
  kind: SubjectKind
  subcategoryId: string
  intent: Intent
  title: string
  author: string
  authorUserId: string | null
  authorUsername: string | null
  at: number
  coordinates: LngLat
  place: string
  description: string
  traits: TraitLine[]
  photoCount: number
}

export interface Subcategory {
  id: string
  categoryId: SubjectKind
  label: string
}

export const categories: { id: SubjectKind; chip: string; label: string; color: string }[] = [
  { id: "item", chip: "Предметы", label: "Предмет", color: "#2f6fed" },
  { id: "person", chip: "Люди", label: "Человек", color: "#c2410c" },
  { id: "animal", chip: "Животные", label: "Животное", color: "#0f766e" },
  { id: "place", chip: "Места", label: "Место", color: "#7c3aed" },
  { id: "event", chip: "События", label: "Событие", color: "#b45309" },
]

export const subcategories: Subcategory[] = [
  { id: "keys", categoryId: "item", label: "Ключи" },
  { id: "electronics", categoryId: "item", label: "Электроника" },
  { id: "bags", categoryId: "item", label: "Сумки" },
  { id: "adults", categoryId: "person", label: "Взрослые" },
  { id: "children", categoryId: "person", label: "Дети" },
  { id: "cats", categoryId: "animal", label: "Кошки" },
  { id: "dogs", categoryId: "animal", label: "Собаки" },
  { id: "yards", categoryId: "place", label: "Дворы" },
  { id: "viewpoints", categoryId: "place", label: "Смотровые" },
  { id: "gatherings", categoryId: "event", label: "Сборы" },
  { id: "meetings", categoryId: "event", label: "Встречи" },
]

export const lossKinds = new Set<SubjectKind>(["item", "person", "animal"])

export function timeCaption(selected: ReadonlySet<SubjectKind>): string | null {
  const loss = [...selected].some((kind) => lossKinds.has(kind))
  const event = selected.has("event")
  if (loss && event) return "Время пропажи/проведения"
  if (loss) return "Время пропажи"
  if (event) return "Время проведения"
  return null
}

export const intents: Record<Intent, string> = {
  seeking: "Ищут",
  found: "Нашли",
  marking: "Отметка",
}

export interface TimeRange {
  from: number | null
  to: number | null
}

export function matchesTime(entry: MapEntry, range: TimeRange, caption: string | null): boolean {
  if (!caption || entry.kind === "place") return true
  if (range.from != null && entry.at < range.from) return false
  if (range.to != null && entry.at > range.to) return false
  return true
}

export function isInside(entry: MapEntry, bounds: LngLatBounds): boolean {
  const [lng, lat] = entry.coordinates
  const [[west, south], [east, north]] = bounds
  return lng >= Math.min(west, east) && lng <= Math.max(west, east) && lat >= Math.min(south, north) && lat <= Math.max(south, north)
}

export function categoryColor(kind: SubjectKind): string {
  return categories.find((category) => category.id === kind)?.color ?? "#333"
}

export const categoryPreset: Record<SubjectKind, string> = {
  item: "islands#blueCircleDotIcon",
  person: "islands#redCircleDotIcon",
  animal: "islands#darkGreenCircleDotIcon",
  place: "islands#violetCircleDotIcon",
  event: "islands#orangeCircleDotIcon",
}

export const categoryPresetSelected: Record<SubjectKind, string> = {
  item: "islands#blueCircleIcon",
  person: "islands#redCircleIcon",
  animal: "islands#darkGreenCircleIcon",
  place: "islands#violetCircleIcon",
  event: "islands#orangeCircleIcon",
}

export function categoryLabel(kind: SubjectKind): string {
  return categories.find((category) => category.id === kind)?.label ?? kind
}

export function subcategoryLabel(id: string): string {
  return subcategories.find((item) => item.id === id)?.label ?? id
}
