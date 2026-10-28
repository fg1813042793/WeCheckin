# Admin visual preview

Run `npm run dev -- --host 127.0.0.1 --port 18085` from `admin/`, then open
`http://127.0.0.1:18085/design-preview.html`.

This standalone Vite entry previews the survey management design using the existing
Vue, Element Plus, Admin UI components and theme tokens. It is not registered in
the production router or menu and is not included in the default production build.
All survey data is synthetic. Search, filters, sorting, pagination, card/list views,
detail/statistics dialogs, CSV export and local create/edit/copy/status/delete actions
work in memory. Reloading the page restores the sample dataset. Sidebar links other
than survey management open the existing admin pages and require normal login.

The preview deliberately retains compact paper cards, with no colored left border,
and limits the page to the existing survey management workflow. Production API,
permissions and deployment integration are deferred until the visual direction is
approved.
