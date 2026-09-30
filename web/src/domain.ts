// Зеркало пакета find/internal/domain для клиента.
// Публичные виды не содержат владельца, скрытую примету и скрытого автора.

export type ID = string

export type Visibility = "public" | "internal"

export type SubjectKind = "item" | "person" | "animal" | "place" | "event"

export type Intent = "seeking" | "found" | "marking"

export type PublicationStatus = "draft" | "published" | "matched" | "closed" | "hidden"

export type Sex = "female" | "male" | "unknown"

export type IdentityStatus = "none" | "pending" | "verified"

export type FollowTarget = "user" | "place" | "event"

export type ExternalSource = "yandex" | "2gis" | "google" | "osm"

export type MatchStatus = "proposed" | "confirmed" | "rejected"

export const TraitHeight = "height_cm"
export const TraitWeight = "weight_kg"
export const TraitNationality = "nationality"
export const TraitAppearance = "appearance"
export const TraitDistinctive = "distinctive"
export const TraitColor = "color"
export const TraitAge = "age"

export interface GeoPoint {
  lat: number
  lon: number
}

export interface PublicUser {
  id: ID
  username: string
  displayName: string
  bio: string
  avatarMediaId: ID | null
  identityStatus: IdentityStatus
}

export interface Trait {
  key: string
  value: string
  visibility: Visibility
}

export interface ItemView {
  id: ID
  name: string
  categoryId: ID
  subcategoryId: ID | null
  description: string
  photoIds: ID[]
  traits: Trait[]
}

export interface PersonView {
  id: ID
  givenName: string
  familyName: string
  patronymic: string
  sex: Sex
  photoIds: ID[]
  traits: Trait[]
}

export interface AnimalView {
  id: ID
  categoryId: ID
  breedCategoryId: ID | null
  breedText: string
  sex: Sex
  nickname: string
  photoIds: ID[]
  traits: Trait[]
}

export interface ExternalRef {
  id: ID
  source: ExternalSource
  externalId: string
  url: string
  snapshot: unknown
}

export interface Place {
  id: ID
  categoryId: ID
  subcategoryId: ID | null
  name: string
  description: string
  address: string
  point: GeoPoint | null
  tags: string[]
  externalRefs: ExternalRef[]
  photoIds: ID[]
  createdBy: ID
  createdAt: string
}

export interface Event {
  id: ID
  categoryId: ID
  subcategoryId: ID | null
  name: string
  description: string
  placeId: ID | null
  address: string
  point: GeoPoint | null
  startsAt: string
  endsAt: string | null
  tags: string[]
  photoIds: ID[]
  createdBy: ID
  createdAt: string
}

export interface PublicationView {
  id: ID
  authorUserId: ID | null
  intent: Intent
  subjectKind: SubjectKind
  subjectId: ID
  body: string
  occurredAt: string | null
  point: GeoPoint | null
  placeLabel: string
  placeId: ID | null
  status: PublicationStatus
  photoIds: ID[]
  createdAt: string
}

export interface Follow {
  followerUserId: ID
  targetKind: FollowTarget
  targetId: ID
  createdAt: string
}

export interface Match {
  id: ID
  seekingPublicationId: ID
  foundPublicationId: ID
  status: MatchStatus
  createdAt: string
}
