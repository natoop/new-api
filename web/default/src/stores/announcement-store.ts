/*
Copyright (C) 2023-2026 QuantumNous

This program is free software: you can redistribute it and/or modify
it under the terms of the GNU Affero General Public License as
published by the Free Software Foundation, either version 3 of the
License, or (at your option) any later version.

This program is distributed in the hope that it will be useful,
but WITHOUT ANY WARRANTY; without even the implied warranty of
MERCHANTABILITY or FITNESS FOR A PARTICULAR PURPOSE. See the
GNU Affero General Public License for more details.

You should have received a copy of the GNU Affero General Public License
along with this program. If not, see <https://www.gnu.org/licenses/>.

For commercial licensing, please contact support@quantumnous.com
*/
import { create } from 'zustand'

interface AnnouncementState {
  // Dialog 可见性
  open: boolean
  // 是否由铃铛手动触发（手动打开忽略 "不再显示" dismiss）
  manual: boolean
  // 受控 onOpenChange：关闭时复位 manual
  setOpen: (open: boolean) => void
  // 点铃铛：强制打开并标记为手动（无视 dismiss）
  openNow: () => void
}

export const useAnnouncementStore = create<AnnouncementState>()((set) => ({
  open: false,
  manual: false,
  setOpen: (open) =>
    set((state) => ({ ...state, open, manual: open ? state.manual : false })),
  openNow: () => set((state) => ({ ...state, open: true, manual: true })),
}))
