<template>
  <v-app>
    <v-main>
      <v-container class="pa-4">
        <v-row>
          <v-col>
            <h1>SQLsmith-Go</h1>

            <v-card class="mb-4" outlined>
              <v-card-title>Health</v-card-title>
              <v-card-text>
                <div v-if="health.error" class="text-error">Error: {{ health.error }}</div>
                <div v-else>
                  <div><strong>Status:</strong> {{ health.status || 'unknown' }}</div>
                  <div v-if="health.timestamp"><strong>Timestamp:</strong> {{ health.timestamp }}</div>
                </div>
              </v-card-text>
              <v-card-actions>
                <v-btn @click="getHealth" variant="outlined">Refresh</v-btn>
              </v-card-actions>
            </v-card>

            <v-card outlined class="mb-4">
              <v-card-title>Info</v-card-title>
              <v-card-text>
                <div v-if="info.error" class="text-error">Error: {{ info.error }}</div>
                <div v-else>
                  <pre style="white-space:pre-wrap">{{ formattedInfo }}</pre>
                </div>
              </v-card-text>
              <v-card-actions>
                <v-btn @click="getInfo" variant="outlined">Refresh</v-btn>
              </v-card-actions>
            </v-card>

            <v-card outlined>
              <v-card-title>Generators</v-card-title>
              <v-card-text>
                <div v-if="generators.error" class="text-error">Error: {{ generators.error }}</div>
                <div v-else>
                  <div><strong>Generator:</strong> {{ generators.generator || 'n/a' }}</div>
                  <div v-if="generators.stmts">
                    <strong>Supported Statements:</strong>
                    <pre style="white-space:pre-wrap">{{ formattedGenerators }}</pre>
                  </div>
                </div>
              </v-card-text>
              <v-card-actions>
                <v-btn @click="getGenerators" variant="outlined">Refresh</v-btn>
              </v-card-actions>
            </v-card>

          </v-col>
        </v-row>
      </v-container>
    </v-main>
  </v-app>
</template>

<script setup>
import { ref, onMounted, computed } from 'vue'

const health = ref({})
const info = ref({})
const generators = ref({})

async function fetchJson(path) {
  const API_BASE = 'http://localhost:8080'
  const url = path.startsWith('http://') || path.startsWith('https://') ? path : `${API_BASE}${path}`
  try {
    const res = await fetch(url, { cache: 'no-store' })
    if (!res.ok) {
      return { error: `HTTP ${res.status} ${res.statusText}` }
    }
    return await res.json()
  } catch (err) {
    return { error: err.message || String(err) }
  }
}

async function getHealth() {
  const data = await fetchJson('/health')
  health.value = data
}

async function getInfo() {
  const data = await fetchJson('/info')
  info.value = data
}

async function getGenerators() {
  const data = await fetchJson('/generators/get')
  generators.value = data
}

const formattedInfo = computed(() => {
  try {
    return JSON.stringify(info.value, null, 2)
  } catch {
    return String(info.value)
  }
})

// Render the generators.stmts map (stmt -> weight) into a readable string
const formattedGenerators = computed(() => {
  const stmts = generators.value && generators.value.stmts ? generators.value.stmts : null
  if (!stmts) return ''

  // Convert to entries and ensure numeric weights
  const entries = Object.entries(stmts).map(([k, v]) => [k, Number(v || 0)])
  const total = entries.reduce((s, [, w]) => s + w, 0)

  // Sort by weight desc, then name
  entries.sort((a, b) => {
    const wdiff = b[1] - a[1]
    if (wdiff !== 0) return wdiff
    return a[0].localeCompare(b[0])
  })

  const lines = entries.map(([k, w]) => {
    const pct = total > 0 ? ((w / total) * 100).toFixed(1) : '0.0'
    return `${k}: ${w} (${pct}%)`
  })

  lines.unshift(`Total tokens: ${total}`)
  return lines.join('\n')
})

onMounted(() => {
  getHealth()
  getInfo()
  getGenerators()
})
</script>

<style scoped>
.text-error { color: #b00020; }
pre { margin: 0; }
</style>
