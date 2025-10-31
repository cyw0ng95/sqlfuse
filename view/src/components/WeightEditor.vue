<template>
  <v-card class="mb-6 modern-card" elevation="8">
    <v-card-title class="d-flex align-center bg-gradient-weights">
      <v-icon class="mr-2">mdi-weight</v-icon>
      Statement Weight Configuration
    </v-card-title>
    <v-card-text class="pa-4">
      <v-alert v-if="!generatorStmts || Object.keys(generatorStmts).length === 0" type="info" variant="tonal" class="mb-4">
        Loading statement types from generator...
      </v-alert>
      <div v-else>
        <div class="mb-4">
          <v-btn color="primary" prepend-icon="mdi-restore" variant="tonal" @click="resetToDefaults" size="small" class="mr-2">
            Reset to Defaults
          </v-btn>
          <v-btn color="secondary" prepend-icon="mdi-content-save" variant="tonal" @click="saveWeights" size="small" class="mr-2">
            Save to Session
          </v-btn>
          <v-btn color="info" prepend-icon="mdi-refresh" variant="tonal" @click="loadWeights" size="small">
            Load from Session
          </v-btn>
        </div>

        <v-text-field
          v-model="searchQuery"
          clearable
          density="compact"
          hide-details
          label="Filter statement types"
          prepend-inner-icon="mdi-magnify"
          variant="outlined"
          class="mb-4"
        />

        <div class="weights-container">
          <v-row v-for="(weight, stmt) in filteredWeights" :key="stmt" class="weight-row align-center mb-2">
            <v-col cols="12" md="5" class="py-1">
              <v-chip size="small" variant="outlined">
                {{ stmt }}
              </v-chip>
            </v-col>
            <v-col cols="12" md="5" class="py-1">
              <v-slider
                v-model="weights[stmt]"
                :max="500"
                :min="0"
                :step="5"
                density="compact"
                hide-details
                thumb-label
                @update:model-value="onWeightChange"
              >
                <template #append>
                  <v-text-field
                    v-model.number="weights[stmt]"
                    density="compact"
                    hide-details
                    single-line
                    style="width: 80px"
                    type="number"
                    variant="outlined"
                    @update:model-value="onWeightChange"
                  />
                </template>
              </v-slider>
            </v-col>
            <v-col cols="12" md="2" class="py-1 text-caption text-right">
              {{ calculatePercentage(stmt) }}%
            </v-col>
          </v-row>
        </div>

        <v-divider class="my-4" />

        <div class="text-subtitle-2 mb-2">
          <strong>Total Weight:</strong> {{ totalWeight }}
        </div>
        <div class="text-caption text-medium-emphasis">
          Changes are automatically saved to session storage. Use "Save to Session" to persist manually.
        </div>
      </div>
    </v-card-text>
  </v-card>
</template>

<script setup>
  import { computed, onMounted, ref, watch } from 'vue'

  const props = defineProps({
    generatorStmts: {
      type: Object,
      default: () => ({})
    }
  })

  const emit = defineEmits(['weights-changed'])

  const weights = ref({})
  const searchQuery = ref('')
  const STORAGE_KEY = 'sqlfuse_custom_weights'

  // Initialize weights from props
  const initializeWeights = () => {
    if (props.generatorStmts && Object.keys(props.generatorStmts).length > 0) {
      weights.value = { ...props.generatorStmts }
      loadWeights() // Load from session storage if available
    }
  }

  // Filter weights based on search query
  const filteredWeights = computed(() => {
    if (!searchQuery.value) {
      return weights.value
    }
    const query = searchQuery.value.toLowerCase()
    return Object.fromEntries(
      Object.entries(weights.value).filter(([key]) =>
        key.toLowerCase().includes(query)
      )
    )
  })

  // Calculate total weight
  const totalWeight = computed(() => {
    return Object.values(weights.value).reduce((sum, w) => sum + (Number(w) || 0), 0)
  })

  // Calculate percentage for a statement type
  const calculatePercentage = (stmt) => {
    const w = weights.value[stmt] || 0
    const total = totalWeight.value
    if (total === 0) return '0.0'
    return ((w / total) * 100).toFixed(1)
  }

  // Reset weights to defaults
  const resetToDefaults = () => {
    if (props.generatorStmts) {
      weights.value = { ...props.generatorStmts }
      saveWeights()
      onWeightChange()
    }
  }

  // Save weights to session storage
  const saveWeights = () => {
    try {
      sessionStorage.setItem(STORAGE_KEY, JSON.stringify(weights.value))
    } catch (error) {
      console.error('Failed to save weights to session storage:', error)
    }
  }

  // Load weights from session storage
  const loadWeights = () => {
    try {
      const saved = sessionStorage.getItem(STORAGE_KEY)
      if (saved) {
        const parsed = JSON.parse(saved)
        // Only merge weights for statement types that exist in current generator
        const validWeights = {}
        for (const [key, value] of Object.entries(parsed)) {
          if (weights.value.hasOwnProperty(key)) {
            validWeights[key] = value
          }
        }
        weights.value = { ...weights.value, ...validWeights }
        onWeightChange()
      }
    } catch (error) {
      console.error('Failed to load weights from session storage:', error)
    }
  }

  // Handle weight change
  const onWeightChange = () => {
    // Auto-save to session storage
    saveWeights()
    // Emit event with updated weights
    emit('weights-changed', weights.value)
  }

  // Watch for generator stmts changes
  watch(() => props.generatorStmts, () => {
    initializeWeights()
  }, { immediate: true })

  onMounted(() => {
    initializeWeights()
  })
</script>

<style scoped>
.modern-card {
  border-radius: 12px !important;
  overflow: hidden;
}

.bg-gradient-weights {
  background: linear-gradient(135deg, #667eea 0%, #764ba2 100%);
  color: white !important;
}

.weights-container {
  max-height: 600px;
  overflow-y: auto;
  padding-right: 8px;
}

.weight-row {
  border-bottom: 1px solid #e0e0e0;
}

.weight-row:last-child {
  border-bottom: none;
}

/* Scrollbar styling */
.weights-container::-webkit-scrollbar {
  width: 8px;
}

.weights-container::-webkit-scrollbar-track {
  background: #f1f1f1;
  border-radius: 4px;
}

.weights-container::-webkit-scrollbar-thumb {
  background: #888;
  border-radius: 4px;
}

.weights-container::-webkit-scrollbar-thumb:hover {
  background: #555;
}
</style>
