<script setup lang="ts">
import '@/assets/styles/pp-mx-home.css'
import { appUpload, updateCurrentTgAvatar, updateCurrentTgName } from '@/api/user'
import { showToast } from 'vant'
import PpmxField from '@/components/PpmxField.vue'
import { usePpmxPageRequest } from '@/composables/usePpmxPageRequest'
import { ppmxProfileFields } from '../ppmx-home/pageData'
import { usePpmxLocale } from '../ppmx-home/composables/usePpmxLocale'
import PpmxPageChrome from '../ppmx-home/components/PpmxPageChrome.vue'

definePage({
  name: 'user-profile',
  meta: {
    titleKey: 'userProfile.title',
    shell: 'ppmx',
    tabbar: false,
    requiresAuth: true,
  },
})

const userStore = useUserStore()
const router = useRouter()
const { t } = useI18n()
const { pickText } = usePpmxLocale()

const savingProfile = ref(false)
const uploadingAvatar = ref(false)
const avatarFileInput = ref<HTMLInputElement>()
const profileForm = reactive({
  username: '',
  firstName: '',
  email: '',
  phone: '',
})

const profileHandle = computed(() => (
  userStore.userInfo?.nickname
  || userStore.userInfo?.username
  || profileForm.username
  || '--'
))
const profileUid = computed(() => {
  const id = userStore.userInfo?.id

  return id ? `UID: ${id}` : 'UID: --'
})
const profileAvatarText = computed(() => {
  const label = profileHandle.value.trim()

  return label && label !== '--' ? label.slice(0, 2).toUpperCase() : 'PP'
})
const profileRequest = usePpmxPageRequest(async () => {
  await userStore.loadUserInfo()
  syncProfileForm()

  return userStore.userInfo ?? null
}, {
  isEmpty: profile => !profile,
  loadingMessage: t('userProfile.loadingTitle'),
})
const profileLoading = computed(() => profileRequest.loading.value)
const profileLoadError = computed(() => profileRequest.hasError.value)

function syncProfileForm() {
  const profile = userStore.userInfo

  profileForm.username = profile?.nickname || profile?.username || ''
  profileForm.firstName = profile?.firstName || profile?.nickname || profile?.username || ''
  profileForm.email = profile?.email || ''
  profileForm.phone = profile?.phone || ''
}

async function loadProfile() {
  await profileRequest.run()
}

function changeAvatar() {
  if (uploadingAvatar.value) return
  avatarFileInput.value?.click()
}

async function handleAvatarFileChange(event: Event) {
  const input = event.target as HTMLInputElement
  const file = input.files?.[0]
  input.value = ''

  if (!file) return

  if (!file.type.startsWith('image/')) {
    showToast(t('userProfile.uploadImageOnly'))
    return
  }

  if (file.size > 5 * 1024 * 1024) {
    showToast(t('userProfile.uploadTooLarge'))
    return
  }

  uploadingAvatar.value = true

  try {
    const uploadResult = await appUpload(file)
    const avatarUrl = uploadResult?.url || ''

    if (!avatarUrl) {
      showToast(t('userProfile.uploadFailed'))
      return
    }

    await updateCurrentTgAvatar(avatarUrl)
    await userStore.loadUserInfo()
    showToast(t('userProfile.avatarUpdated'))
  } finally {
    uploadingAvatar.value = false
  }
}

async function saveProfile() {
  if (savingProfile.value) return

  const nextName = profileForm.username.trim()

  if (!nextName) {
    showToast(t('userProfile.nameRequired'))
    return
  }

  if ([...nextName].length > 64) {
    showToast(t('userProfile.nameTooLong'))
    return
  }

  savingProfile.value = true

  try {
    await updateCurrentTgName(nextName)
    await userStore.loadUserInfo()
    showToast(t('userProfile.nameUpdated'))
  } finally {
    savingProfile.value = false
  }
}

function backToAccount() {
  void router.push('/profile')
}

onMounted(() => {
  if (userStore.userInfo) {
    syncProfileForm()
    return
  }

  void loadProfile()
})

watch(
  () => userStore.userInfo,
  (userInfo) => {
    if (userInfo) syncProfileForm()
  },
)
</script>

<template>
  <main class="ppmx-page ppmx-subpage ppmx-user-profile-page">
    <PpmxPageChrome>
      <section class="ppmx-subpage__compact ppmx-user-profile-stack">
        <header class="ppmx-subpage-head ppmx-user-profile-head reveal">
          <div class="ppmx-eyebrow">{{ t('userProfile.eyebrow') }}</div>
          <h1 class="ppmx-display">{{ t('userProfile.title') }}</h1>
          <p>{{ t('userProfile.description') }}</p>
        </header>

        <AppState
          v-if="profileLoading"
          class="ppmx-account-state"
          status="loading"
          :title="t('userProfile.loadingTitle')"
          :message="t('userProfile.loadingMessage')"
          skeleton-variant="user-profile-page"
        />

        <AppState
          v-else-if="profileLoadError"
          class="ppmx-account-state"
          status="error"
          :title="t('userProfile.loadErrorTitle')"
          :message="t('userProfile.loadErrorMessage')"
          :action-text="t('common.retry')"
          @retry="loadProfile"
        />

        <template v-else>
          <section class="ppmx-user-profile-card ppmx-user-profile-identity reveal">
            <div class="ppmx-user-profile-avatar-wrap">
              <span id="profileAvatar" class="ppmx-user-profile-avatar">
                <img v-if="userStore.userInfo?.avatar" :src="userStore.userInfo.avatar" alt="" />
                <span v-else>{{ profileAvatarText }}</span>
              </span>
              <button type="button" :aria-label="t('userProfile.changeAvatar')" @click="changeAvatar">
                <i class="fa-solid" :class="uploadingAvatar ? 'fa-spinner fa-spin' : 'fa-camera'" />
              </button>
              <input
                ref="avatarFileInput"
                class="ppmx-user-profile-file"
                type="file"
                accept="image/*"
                @change="handleAvatarFileChange"
              />
            </div>

            <div class="ppmx-user-profile-summary">
              <p>{{ t('userProfile.username') }}</p>
              <strong id="profileHandle">
                <i class="fa-solid fa-circle-user" />
                {{ profileHandle }}
              </strong>
              <small>{{ profileUid }}</small>

              <PpmxField field-class="ppmx-user-profile-username" :label="t('userProfile.usernameInput')">
                <input id="profileUserInput" v-model.trim="profileForm.username" type="text" />
              </PpmxField>
            </div>
          </section>

          <section class="ppmx-user-profile-card ppmx-user-profile-details reveal">
            <div class="ppmx-user-profile-details__head">
              <div>
                <div class="ppmx-eyebrow">{{ t('userProfile.detailsEyebrow') }}</div>
                <h2>{{ t('userProfile.detailsTitle') }}</h2>
              </div>
              <i class="fa-solid fa-id-card" />
            </div>

            <form id="profileDetailsForm" class="ppmx-user-profile-form" @submit.prevent="saveProfile">
              <div class="ppmx-user-profile-grid">
                <PpmxField
                  v-for="field in ppmxProfileFields"
                  :key="field.id"
                  field-class="ppmx-user-profile-field"
                  :label="pickText(field.label)"
                >
                  <input
                    v-model.trim="profileForm[field.model]"
                    :type="field.type || 'text'"
                    autocomplete="off"
                  />
                </PpmxField>
              </div>

              <div class="ppmx-user-profile-actions">
                <button class="ppmx-btn-ghost" type="button" @click="backToAccount">
                  {{ t('userProfile.backAccount') }}
                </button>
                <button class="ppmx-btn-red" type="submit" :disabled="savingProfile">
                  {{ savingProfile ? t('userProfile.saving') : t('userProfile.save') }}
                </button>
              </div>
            </form>
          </section>
        </template>
      </section>
    </PpmxPageChrome>
  </main>
</template>
