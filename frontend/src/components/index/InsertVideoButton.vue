<script setup lang="ts">
import UserInputToolButton from "./UserInputToolButton.vue"
import {ref} from "vue"
import {GetYoutubeTranscript, GetYoutubeVideo} from "../../../wailsjs/go/main/App"
import {main, util} from "../../../wailsjs/go/models"
import {swal} from "../../helper"
import YoutubeVideoResult = main.YoutubeVideoResult
import YtCustomCaption = util.YtCustomCaption

let props = defineProps<{
  isAsking: boolean,
}>()
let emit = defineEmits<{
  (e: 'append', text: string): void
}>()
let dialog = ref(false)
let youtubeLink = ref('')
let youtubeResult = ref<undefined | YoutubeVideoResult>(undefined)
let loading = ref(false)

function handleInsert(caption: YtCustomCaption) {
  loading.value = true
  GetYoutubeTranscript(caption).then(res => {
    let obj: any = {
      ...youtubeResult.value?.details,
      'transcript': res,
      'transcript_lang': caption.language_code,
    }
    emit('append', '[user](#youtube_video_info)\n' + JSON.stringify(obj) + '\n\n')
    dialog.value = false
    youtubeLink.value = ''
    youtubeResult.value = undefined
  }).catch(err => {
    swal.error(err)
  }).finally(() => {
    loading.value = false
  })
}

function handleFetch() {
  loading.value = true
  youtubeResult.value = undefined
  GetYoutubeVideo(youtubeLink.value).then(res => {
    youtubeResult.value = res
  }).catch(err => {
    swal.error(err)
  }).finally(() => {
    loading.value = false
  })
}
</script>

<template>
  <div>
    <user-input-tool-button icon="mdi-video" :disabled="isAsking" @click="dialog=true"
                            tooltip="Insert a Youtube Video to chat context"></user-input-tool-button>
    <v-dialog max-width="500" v-model="dialog" :scrollable="true">
      <v-card title="Insert a Youtube Video">
        <v-card-text>
          <div class="d-flex align-center">
            <v-text-field @keydown.enter="handleFetch" density="compact" v-model="youtubeLink" label="Youtube Link"
                          color="primary"></v-text-field>
            <v-btn color="primary" class="ml-3 mb-4" :loading="loading" variant="tonal" @click="handleFetch">Fetch
            </v-btn>
          </div>
          <div v-if="youtubeResult">
            <div style="font-size: 17px;font-weight: bold">Video Details</div>
            <p>Title: <i>{{ youtubeResult.details.title }}</i></p>
            <p>Author: <i>{{ youtubeResult.details.author }}</i></p>
            <p>Cover:</p>
            <v-img :src="youtubeResult.details.pic_url"></v-img>
            <div style="font-size: 17px;font-weight: bold" class="mt-3">Select a Caption</div>
            <v-list density="compact">
              <v-list-item v-for="caption in youtubeResult.captions">
                <template #title>
                  <div class="d-flex align-center">
                    <p>{{ caption.name }}</p>
                    <v-chip class="ml-3" size="small" color="#5c5c5c" v-if="caption.is_asr">Auto-generated</v-chip>
                    <v-chip class="ml-3" size="small" color="#2b28ec" v-else-if="caption.is_translated">Translated
                    </v-chip>
                    <v-chip class="ml-3" size="small" color="green" v-else>Manually-inputted</v-chip>
                    <v-spacer></v-spacer>
                    <v-btn density="compact" variant="tonal" color="primary" @click="handleInsert(caption)"
                           :loading="loading">Insert
                    </v-btn>
                  </div>
                </template>
              </v-list-item>
            </v-list>
          </div>
        </v-card-text>
        <v-card-actions>
          <v-spacer></v-spacer>
          <v-btn variant="text" color="primary" @click="dialog=false">Cancel</v-btn>
        </v-card-actions>
      </v-card>
    </v-dialog>
  </div>
</template>

<style scoped>

</style>