<template>
  <v-card outlined class="mb-4">
    <v-card-title>Jobs</v-card-title>
    <v-card-text>
      <v-row>
        <v-col cols="12" sm="8">
          <v-select
            v-model="selectedExecutor"
            :items="executors"
            item-title="executor"
            item-value="executor"
            label="Executor"
            :disabled="executors.length === 0"
            hide-details
          />
          <v-text-field v-model="args" label="Arguments (separated by space)" placeholder="--workers 4 --queries 100" class="mt-2" />
        </v-col>
        <v-col cols="12" sm="4" class="d-flex align-center">
          <v-btn @click="createJob" :loading="creating" variant="contained">Start Job</v-btn>
        </v-col>
      </v-row>

      <div v-if="createError" class="text-error">Error: {{ createError }}</div>
      <div v-if="lastID">Started job id: <strong>{{ lastID }}</strong></div>

      <v-divider class="my-4"></v-divider>

      <v-row>
        <v-col cols="12" sm="6">
          <v-text-field v-model="queryID" label="Job ID" />
        </v-col>
        <v-col cols="12" sm="6" class="d-flex align-center">
          <v-btn @click="fetchStatus" variant="outlined">Status</v-btn>
          <v-btn @click="fetchInfo" class="ml-2" variant="outlined">Info</v-btn>
          <v-btn @click="stopJob" class="ml-2" color="error" variant="outlined">Stop</v-btn>
        </v-col>
      </v-row>

      <div v-if="status"> <strong>Status:</strong> {{ status.status }} <span v-if="status.pid">(pid: {{ status.pid }})</span></div>
      <div v-if="status && status.started_at">Started: {{ status.started_at }}</div>
      <div v-if="status && status.ended_at">Ended: {{ status.ended_at }}</div>
      <div v-if="status && status.exit_code != null">Exit: {{ status.exit_code }}</div>

      <div v-if="infoData" class="mt-3">
        <strong>Stdout:</strong>
        <pre style="white-space:pre-wrap">{{ infoData.stdout }}</pre>
        <strong>Stderr:</strong>
        <pre style="white-space:pre-wrap">{{ infoData.stderr }}</pre>
      </div>

    </v-card-text>
    <v-card-actions>
      <v-btn @click="clear" variant="text">Clear</v-btn>
    </v-card-actions>
  </v-card>
</template>

<script setup>
import { ref, onMounted } from 'vue'

const API_BASE = 'http://localhost:8080'

const executors = ref([])
const selectedExecutor = ref('')
const args = ref('')
const creating = ref(false)
const createError = ref('')
const lastID = ref('')

const queryID = ref('')
const status = ref(null)
const infoData = ref(null)

async function fetchJson(path, opts) {
  const url = path.startsWith('http://') || path.startsWith('https://') ? path : `${API_BASE}${path}`
  try {
    const res = await fetch(url, opts)
    if (!res.ok) {
      const text = await res.text()
      return { error: `HTTP ${res.status} ${res.statusText}: ${text}` }
    }
    return await res.json()
  } catch (err) {
    return { error: err.message || String(err) }
  }
}

async function getExecutors() {
  const data = await fetchJson('/executors')
  if (data && data.error) {
    executors.value = []
    return
  }
  // data expected to be an array of { executor, path }
  executors.value = Array.isArray(data) ? data : []
}

onMounted(() => {
  getExecutors()
})

async function createJob() {
  createError.value = ''
  if (!selectedExecutor.value) {
    createError.value = 'executor is required'
    return
  }
  creating.value = true
  // split args by whitespace
  const argsList = args.value.trim() === '' ? [] : (args.value.trim().match(/\S+/g) || [])
  const payload = { executor: selectedExecutor.value, args: argsList }
  const data = await fetchJson('/job/new', { method: 'POST', headers: { 'Content-Type': 'application/json' }, body: JSON.stringify(payload) })
  creating.value = false
  if (data.error) {
    createError.value = data.error
    return
  }
  lastID.value = data.id || ''
  queryID.value = lastID.value
}

async function fetchStatus() {
  if (!queryID.value) return
  const data = await fetchJson(`/job/status?id=${encodeURIComponent(queryID.value)}`)
  if (data.error) {
    status.value = { error: data.error }
    return
  }
  status.value = data
}

async function fetchInfo() {
  if (!queryID.value) return
  const data = await fetchJson(`/job/info?id=${encodeURIComponent(queryID.value)}`)
  if (data.error) {
    infoData.value = { error: data.error }
    return
  }
  infoData.value = data
}

async function stopJob() {
  if (!queryID.value) return
  const data = await fetchJson(`/job/stop?id=${encodeURIComponent(queryID.value)}`, { method: 'POST' })
  if (data.error) {
    status.value = { error: data.error }
    return
  }
  // refresh status
  await fetchStatus()
}

function clear() {
  selectedExecutor.value = ''
  args.value = ''
  lastID.value = ''
  queryID.value = ''
  status.value = null
  infoData.value = null
  createError.value = ''
}
</script>

<style scoped>
.text-error { color: #b00020; }
pre { margin: 0; }
</style>
