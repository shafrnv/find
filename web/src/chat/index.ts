/** Блок чата: верификация по отличительной особенности и переписка.
 *  Встреча (meeting) — отдельный модуль, сюда не входит. */
export { ChatsPanel } from "./ChatsPanel"
export { claimFound, loadDialog, loadDialogs, sendDialogMessage, submitDialogAnswer, confirmDialogAnswer } from "./api"
export type { DialogDetail, DialogSummary, ChatMessage, DialogStatus } from "./api"
