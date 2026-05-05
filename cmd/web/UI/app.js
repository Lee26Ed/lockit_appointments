import { emitter }     from './modules/event-emitter.js'
import { DataService } from './modules/data-service.js'

// DataService.setAuthToken('DVXFATM4DTOU53A5AV5CNY7M5M');
  
// ── 1. State — single source of truth ──
const state = { 
               services: [],
               loading: false,
               error: null,
               currentPage: 1,
               pageSize: 5,
               metadata: {
                 current_page: 1,
                 page_size: 5,
                 first_page: 1,
                 last_page: 1,
                 total_records: 0
               },
              };

function render() {
  const tableBody = document.querySelector('#table-body')
  const paginationControls = document.querySelector('#pagination-controls')

  if (!tableBody || !paginationControls) return

  if (state.loading) {
    tableBody.innerHTML = `<tr><td colspan="4" style="text-align: center; padding: 40px;">Loading...</td></tr>`
    paginationControls.innerHTML = ''
    return
  }

  if (state.error) {
    tableBody.innerHTML = `<tr><td colspan="4" style="text-align: center; padding: 40px; color: red;">⚠ ${state.error}</td></tr>`
    paginationControls.innerHTML = ''
    return
  }

  if (state.services.length === 0) {
    tableBody.innerHTML = `<tr><td colspan="4" style="text-align: center; padding: 40px;">No services yet.</td></tr>`
    paginationControls.innerHTML = ''
    return
  }

  // Use API metadata for pagination
  const totalPages = state.metadata.last_page
  const paginatedServices = state.services

  // Render table rows
  tableBody.innerHTML = paginatedServices
    .map(s => `<tr>
      <td>${s.name || 'N/A'}</td>
      <td>${s.business_name || 'N/A'}</td>
      <td>$${s.price || 'N/A'}</td>
      <td>${s.duration || 'N/A'} min</td>
    </tr>`)
    .join('')

  // Render pagination controls
  let paginationHTML = `
    <button id="prev-btn" ${state.currentPage === 1 ? 'disabled' : ''}>← Previous</button>
    <span class="page-info">Page ${state.currentPage} of ${totalPages}</span>
    <button id="next-btn" ${state.currentPage === totalPages ? 'disabled' : ''}>Next →</button>
  `
  paginationControls.innerHTML = paginationHTML

  // Add event listeners to pagination buttons
  const prevBtn = document.querySelector('#prev-btn')
  const nextBtn = document.querySelector('#next-btn')

  if (prevBtn) prevBtn.addEventListener('click', () => {
    if (state.currentPage > 1) {
      state.currentPage--
      DataService.fetchServices(state.currentPage, state.pageSize)
    }
  })

  if (nextBtn) nextBtn.addEventListener('click', () => {
    if (state.currentPage < totalPages) {
      state.currentPage++
      DataService.fetchServices(state.currentPage, state.pageSize)
    }
  })
}

// ── 3. Observers — update state, then call render() ───
  emitter.on('services:loading', () => {
  state.loading = true
  state.error   = null
  render()
  })
 
  emitter.on('services:loaded', (data) => {
  state.services = data.services
  state.metadata = data.metadata
  state.currentPage = data.metadata.current_page
  state.loading = false
  render()
})

emitter.on('services:error', (msg) => {
  state.error   = msg
  state.loading = false
  render()
  })


    // ── 4. Boot ──
   render()                    // initial paint (empty state)
   DataService.fetchServices()   // kick off the first fetch