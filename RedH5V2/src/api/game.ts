import { get, post } from '@/utils/http'

export interface AppHomeGameItem {
  gameId: number
  gameName: string
  categoryCode: string
  type?: number
  manufacturer?: string
  gameIcon?: string
  horizontalImage?: string
  sort?: number
  showIndex?: number
}

export type AppHomeGameData = Record<string, AppHomeGameItem[] | undefined>

export interface AppGameListReq {
  currentPage: number
  pageSize: number
  categoryCode?: string
  thirdGameCategory?: string
  gameName?: string
}

export interface AppGameListResp {
  list: AppHomeGameItem[]
  total: number
  pageSize: number
  currentPage: number
}

export interface AppGameCategoryListResp extends AppGameListResp {
  thirdGameCategories?: string[]
  thirdGameCategory?: string
}

export interface AppGameThirdCategoryResp {
  list: string[]
}

export interface AppGameLaunchReq {
  gameId: number
  language: string
}

export interface AppGameLaunchResp {
  url?: string
  content?: string
}

export function getAppGameHome() {
  return get<AppHomeGameData>('/v1/app/appGame/home')
}

export function getAppGameList(data: AppGameListReq) {
  return post<AppGameListResp>('/v1/app/appGame/list', data)
}

export function getAppGameCategoryList(data: AppGameListReq) {
  return post<AppGameCategoryListResp>('/v1/app/appGame/categoryList', data)
}

export function getAppGameThirdCategories(data: { categoryCode?: string }) {
  return post<AppGameThirdCategoryResp>('/v1/app/appGame/thirdCategories', data)
}

export function launchAppGame(data: AppGameLaunchReq) {
  return post<AppGameLaunchResp>('/v1/app/appGame/launch', data)
}

export function launchGscSportGame() {
  return post<AppGameLaunchResp>('/v1/app/gsc/launch', {})
}
