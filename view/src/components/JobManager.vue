<template>
  <v-card class="mb-6 modern-card" elevation="8">
    <v-card-title class="d-flex align-center bg-gradient-jobs">
      <v-icon class="mr-2">mdi-briefcase</v-icon>
      Job Management
    </v-card-title>
    <v-card-text class="pa-4">
      <v-row>
        <v-col cols="12" md="8">
          <v-select
            v-model="selectedExecutor"
            :items="executors"
            item-title="executor"
            item-value="executor"
            label="Executor"
            :disabled="executors.length === 0"
            variant="outlined"
            density="comfortable"
            prepend-inner-icon="mdi-cog"
            hide-details
          >
            <template v-slot:item="{ props, item }">
              <v-list-item v-bind="props">
                <template v-slot:prepend>
                  <v-icon>mdi-engine</v-icon>
                </template>
                <template v-slot:title>
                  {{ item.raw.executor }}
                </template>
                <template v-slot:subtitle v-if="item.raw.flavor">
                  Flavor: {{ item.raw.flavor }}
                </template>
              </v-list-item>
            </template>
            <template v-slot:selection="{ item }">
              <span>{{ item.raw.executor }}</span>
              <span v-if="item.raw.flavor" class="text-caption ml-2 text-primary">({{ item.raw.flavor }})</span>
            </template>
          </v-select>
          <v-text-field 
            v-model="args" 
            label="Arguments (separated by space)" 
            placeholder="--workers 4 --queries 100" 
            class="mt-3"
            variant="outlined"
            density="comfortable"
            prepend-inner-icon="mdi-code-tags"
            hide-details
          />
        </v-col>
        <v-col cols="12" md="4" class="d-flex align-center">
          <v-btn 
            @click="createJob" 
            :loading="creating" 
            color="primary"
            size="large"
            block
            prepend-icon="mdi-play-circle"
            elevation="2"
          >
            Start Job
          </v-btn>
        </v-col>
      </v-row>

      <div v-if="createError" class="error-alert mt-3">
        <v-icon color="error" class="mr-2">mdi-alert-circle</v-icon>
        {{ createError }}
      </div>
      <div v-if="lastID" class="success-alert mt-3">
        <v-icon color="success" class="mr-2">mdi-check-circle</v-icon>
        Started job ID: <strong>{{ lastID }}</strong>
      </div>

      <v-divider class="my-6"></v-divider>

      <div class="text-subtitle-1 font-weight-bold mb-3">
        <v-icon class="mr-2">mdi-magnify</v-icon>
        Query Job Status
      </div>
      
      <v-row>
        <v-col cols="12" md="6">
          <v-text-field 
            v-model="queryID" 
            label="Job ID" 
            variant="outlined"
            density="comfortable"
            prepend-inner-icon="mdi-identifier"
            hide-details
          />
        </v-col>
        <v-col cols="12" md="6" class="d-flex align-center gap-2">
          <v-btn @click="fetchStatus" color="info" variant="tonal" prepend-icon="mdi-information">
            Status
          </v-btn>
          <v-btn @click="fetchInfo" color="info" variant="tonal" prepend-icon="mdi-text-box">
            Info
          </v-btn>
          <v-btn @click="stopJob" color="error" variant="tonal" prepend-icon="mdi-stop-circle">
            Stop
          </v-btn>
        </v-col>
      </v-row>

      <div v-if="status" class="status-box mt-4">
        <v-chip color="info" class="mb-2" prepend-icon="mdi-information">
          {{ status.status }}
          <span v-if="status.pid" class="ml-2">(PID: {{ status.pid }})</span>
        </v-chip>
        <div v-if="status.started_at" class="text-caption">
          <v-icon size="small">mdi-clock-start</v-icon> Started: {{ status.started_at }}
        </div>
        <div v-if="status.ended_at" class="text-caption">
          <v-icon size="small">mdi-clock-end</v-icon> Ended: {{ status.ended_at }}
        </div>
        <div v-if="status.exit_code != null" class="text-caption">
          <v-icon size="small">mdi-exit-to-app</v-icon> Exit Code: {{ status.exit_code }}
        </div>
      </div>

      <div v-if="infoData" class="mt-4">
        <v-expansion-panels>
          <v-expansion-panel>
            <v-expansion-panel-title>
              <v-icon class="mr-2">mdi-console</v-icon>
              Standard Output
            </v-expansion-panel-title>
            <v-expansion-panel-text>
              <pre class="output-pre">{{ infoData.stdout || '(empty)' }}</pre>
            </v-expansion-panel-text>
          </v-expansion-panel>
          <v-expansion-panel>
            <v-expansion-panel-title>
              <v-icon class="mr-2">mdi-alert</v-icon>
              Standard Error
            </v-expansion-panel-title>
            <v-expansion-panel-text>
              <pre class="output-pre error-output">{{ infoData.stderr || '(empty)' }}</pre>
            </v-expansion-panel-text>
          </v-expansion-panel>
        </v-expansion-panels>
      </div>

    </v-card-text>
    <v-card-actions class="pa-4 pt-0">
      <v-btn @click="clear" color="secondary" variant="text" prepend-icon="mdi-broom">
        Clear
      </v-btn>
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
.modern-card {
  border-radius: 12px !important;
  overflow: hidden;
}

.bg-gradient-jobs {
  background: linear-gradient(135deg, #f093fb 0%, #f5576c 100%);
  color: white !important;
}

.error-alert {
  color: #d32f2f;
  background-color: #ffebee;
  padding: 12px 16px;
  border-radius: 8px;
  border-left: 4px solid #d32f2f;
  display: flex;
  align-items: center;
}

.success-alert {
  color: #2e7d32;
  background-color: #e8f5e9;
  padding: 12px 16px;
  border-radius: 8px;
  border-left: 4px solid #4caf50;
  display: flex;
  align-items: center;
}

.status-box {
  background-color: #f5f5f5;
  padding: 16px;
  border-radius: 8px;
  border: 1px solid #e0e0e0;
}

.output-pre {
  background-color: #263238;
  color: #aed581;
  border-radius: 8px;
  padding: 16px;
  overflow-x: auto;
  font-family: 'Courier New', monospace;
  font-size: 0.875rem;
  white-space: pre-wrap;
  margin: 0;
}

.error-output {
  color: #ef9a9a;
}

.gap-2 {
  gap: 8px;
}
</style>
