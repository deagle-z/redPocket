import askmeslotIcon from '@/assets/images/game_tabs/providers/slot-askme.png'
import bgIcon from '@/assets/images/game_tabs/providers/slot-bgaming.png'
import bngIcon from '@/assets/images/game_tabs/providers/slot-bng.png'
import btgIcon from '@/assets/images/game_tabs/providers/slot-btg.png'
import cpIcon from '@/assets/images/game_tabs/providers/slot-cp.png'
import cq9Icon from '@/assets/images/game_tabs/providers/slot-cq9.png'
import evoIcon from '@/assets/images/game_tabs/providers/slot-evoplay.png'
import fachaiIcon from '@/assets/images/game_tabs/providers/slot-fachai.png'
import g759Icon from '@/assets/images/game_tabs/providers/slot-g759.png'
import habaneroIcon from '@/assets/images/game_tabs/providers/slot-habanero.png'
import hacksawIcon from '@/assets/images/game_tabs/providers/slot-hacksaw.png'
import inoutIcon from '@/assets/images/game_tabs/providers/slot-inout.png'
import jdbIcon from '@/assets/images/game_tabs/providers/slot-jdb.png'
import jiliIcon from '@/assets/images/game_tabs/providers/slot-jili.png'
import kaIcon from '@/assets/images/game_tabs/providers/slot-ka.png'
import mascotIcon from '@/assets/images/game_tabs/providers/slot-mascot.png'
import mgIcon from '@/assets/images/game_tabs/providers/slot-mg.png'
import netentIcon from '@/assets/images/game_tabs/providers/slot-netent.png'
import oaksIcon from '@/assets/images/game_tabs/providers/slot-3oaks.png'
import pgIcon from '@/assets/images/game_tabs/providers/slot-pgsoft.png'
import playtechIcon from '@/assets/images/game_tabs/providers/slot-playtech.png'
import popiplayIcon from '@/assets/images/game_tabs/providers/slot-popiplay.png'
import popokIcon from '@/assets/images/game_tabs/providers/slot-popOK.png'
import ppIcon from '@/assets/images/game_tabs/providers/slot-pragmatic.png'
import r88Icon from '@/assets/images/game_tabs/providers/slot-r88.png'
import rtIcon from '@/assets/images/game_tabs/providers/slot-redtiger.png'
import spadeIcon from '@/assets/images/game_tabs/providers/slot-spade.png'
import spribeIcon from '@/assets/images/game_tabs/providers/spribe.png'
import tadaIcon from '@/assets/images/game_tabs/providers/slot-tada.png'
import wgIcon from '@/assets/images/game_tabs/providers/slot-wg.png'

export const gameProviderIcons: Record<string, string> = {
  askmeslot: askmeslotIcon,
  bg: bgIcon,
  bng: bngIcon,
  btg: btgIcon,
  cp: cpIcon,
  cq9: cq9Icon,
  evo: evoIcon,
  fachai: fachaiIcon,
  g759: g759Icon,
  habanero: habaneroIcon,
  hacksaw: hacksawIcon,
  inout: inoutIcon,
  jdb: jdbIcon,
  jili: jiliIcon,
  ka: kaIcon,
  mascot: mascotIcon,
  mg: mgIcon,
  netent: netentIcon,
  oaks: oaksIcon,
  pg: pgIcon,
  playtech: playtechIcon,
  popiplay: popiplayIcon,
  popok: popokIcon,
  pp: ppIcon,
  r88: r88Icon,
  rt: rtIcon,
  spade: spadeIcon,
  spribe: spribeIcon,
  tada: tadaIcon,
  wg: wgIcon,
}

export function normalizeGameProviderCode(code?: string | null) {
  return String(code || '').trim().toLowerCase()
}

export function getGameProviderIcon(code?: string | null) {
  return gameProviderIcons[normalizeGameProviderCode(code)]
}

export function getGameProviderLabel(code?: string | null) {
  return normalizeGameProviderCode(code).toUpperCase() || 'GAME'
}
