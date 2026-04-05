# _mavubit Component Usage Patterns

How the shared `_mavubit` library components are used across 5 Elixir/Phoenix projects.

---

## Architecture: _mavubit → _bit → Project Code

The `_mavubit` layer is shared source code (managed via bit.dev). Project code rarely references `_mavubit` modules directly. Instead, a `_bit/` directory inside `lib/my_app_be/` re-exports modules under the project namespace:

```
lib/_mavubit/be1/be/mavu_list/...       ← shared source (MavuBe1.Be.MavuList.*)
lib/my_app_be/_bit/mavu_datagrid/...    ← re-exported as MyAppBe.MavuDatagrid
lib/my_app_be/_bit/mavu_subforms/...    ← re-exported as MyAppBe.MavuSubforms
lib/my_app_be/_bit/mavu_list_ashfilter/ ← re-exported as MyAppBe.MavuListAshfilter
```

**In project code, prefer `MyAppBe.*` (from `_bit/`)** for UI components like `MavuDatagrid`, `MavuSubforms`, `MavuListAshfilter`. Use the direct `Mavu*` namespace for utility functions like `MavuUtils`, `MavuBuckets`, `MavuEntities`.

### Shared _mavubit components (by adoption)

| Component | Projects |
|-----------|----------|
| `essentials` | 5/5 |
| `be1` | 5/5 |
| `fe1` | 3/5 |
| `sulu2` | 3/5 |
| `kiosk1` | 3/5 |

### Shared _bit re-exports (by adoption)

| _bit module | Projects |
|-------------|----------|
| `mavu_datagrid` | 4/5 |
| `mavu_subforms` | 4/5 |
| `mavu_list_ashfilter` | 4/5 |
| `mavu_nav` | 4/5 |
| `mavu_tags_live` | 4/5 |
| `mavu_tagcats_live` | 4/5 |
| `flightcontrol_live` | 4/5 |
| `joblogs_live` | 4/5 |
| `obanjobs_live` | 4/5 |
| `tagged_rodents_live` | 4/5 |

---

## MavuUtils — Universal Helpers (5/5 projects, 281 files)

The most-used shared module. Common functions by frequency:

| Function | Calls | Purpose |
|----------|-------|---------|
| `update_params_in_url/2` | 318 | Patch URL query params for navigation |
| `present?/1` | 199 | Nil/empty guard |
| `log/1` | 191 | Debug logging |
| `to_int/1` | 62 | Safe integer conversion |
| `empty?/1` | 56 | Nil/empty check |
| `any_get/2` | 40 | Get from map/struct/keyword |
| `do_if_present/2` | 37 | Conditional transform |
| `if_empty/2` | 34 | Default when empty |
| `if_nil/2` | 30 | Default when nil |
| `string_wrap/1` | 28 | Ensure string type |
| `to_integer/1` | 20 | Integer conversion variant |
| `map_wrap/1` | 15 | Ensure map type |
| `die/2` | 14 | Debug halt (dev only) |
| `true?/1` | 10 | Truthy check |

**Key pattern — URL state management:**
```elixir
push_patch(to: MavuUtils.update_params_in_url(socket.assigns.context.current_url,
  rec: result.id,
  tab: "main"
))
```

---

## MavuResourceLive — Scaffolded CRUD (3/5 projects)

A three-piece scaffold for standard list/edit screens. Used in wzz-ex, chatty, and filmarchiv-ex. Projects import base modules and pass an `@optsmap` config.

### Shell (`TopLvBase` + `TopLvThemeAlphaC`)

```elixir
defmodule MyAppBe.ThingsLive do
  use MyAppBe, :live_view
  import MavuBe1.Be.MavuResourceLive.TopLvBase

  @optsmap %{resource: MyApp.Thing, top_lv_module: __MODULE__}

  def mount(params, session, socket) do
    {:ok, base_mount(params, session, socket, @optsmap)}
  end

  def handle_params(params, _url, socket) do
    {:noreply, apply_action(socket, socket.assigns.live_action, params)}
  end

  def render(assigns) do
    ~H"""
    <MavuBe1.Be.MavuResourceLive.TopLvThemeAlphaC.full {assigns}
      edit_live_component={MyAppBe.ThingsLive.ThingEditComponentLc}
      list_live_component={MyAppBe.ThingsLive.ThingListLc}
      mode={:record}
    />
    """
  end

  def handle_event(event, msg, socket), do: base_handle_event(event, msg, socket)
  def handle_info(msg, socket), do: base_handle_info(msg, socket)
end
```

### List (`ListLcBase` + `ListLcThemeAlphaC`)

```elixir
defmodule MyAppBe.ThingsLive.ThingListLc do
  use MyAppBe, :live_component
  import MavuBe1.Be.MavuResourceLive.ListLcBase

  @optsmap %{resource: MyApp.Thing, list_lv_module: __MODULE__}

  def update(assigns, socket) do
    {:ok, base_update(socket, assigns, @optsmap) |> load_items()}
  end

  def load_items(socket), do: base_load_items(socket)
  def listconf(assigns), do: base_listconf(assigns)
  def listfilter(source, conf, tweaks), do: base_listfilter(source, conf, tweaks)
  def handle_event(event, msg, socket), do: base_handle_event(event, msg, socket)

  def render(assigns) do
    ~H"""
    <MavuBe1.Be.MavuResourceLive.ListLcThemeAlphaC.full {assigns}
      headline_str="Things"
      expandable_rows?={true}
      row_btn_edit?={true}
      detail_btn_delete?={true}
    />
    """
  end
end
```

### Edit (`EditLcBase` + `EditLcThemeAlphaC`)

```elixir
defmodule MyAppBe.ThingsLive.ThingEditComponentLc do
  use MyAppBe, :live_component
  import MavuBe1.Be.MavuResourceLive.EditLcBase

  @optsmap %{
    resource: MyApp.Thing,
    tabconf: [%{key: :main, title: "Main"}],
    edit_lv_module: __MODULE__
  }

  def update(assigns, socket), do: {:ok, base_update(assigns, socket, @optsmap)}
  def load_item(assigns, subforms), do: base_load_item(assigns, subforms)
  def handle_event(event, msg, socket), do: base_handle_event(event, msg, socket)

  def render(assigns) do
    ~H"""
    <MavuBe1.Be.MavuResourceLive.EditLcThemeAlphaC.full {assigns}>
      <%= case @tabs.current do %>
        <% :main -> %>
          <.ash_input field={@form[:name]} />
          <.ash_input field={@form[:description]} />
      <% end %>
    </MavuBe1.Be.MavuResourceLive.EditLcThemeAlphaC.full>
    """
  end
end
```

**Note:** climate and eselrw build CRUD screens manually (no scaffold), wiring `MavuListAsh`, `MavuDatagrid`, and `AshPhoenix.Form` directly.

---

## MavuDatagrid — Data Tables (5/5 projects, 64 files)

Referenced as `MyAppBe.MavuDatagrid.datagrid_c` (via `_bit/`). Uses typed attrs and HEEx slots.

**Primary slots:**
- `<:col>` — column content (229 uses)
- `<:colheader_style>` — column header styling (96 uses)

**Key attrs:** `id`, `rows`, `visible_cols`, `first_cols`, `last_cols`, `skip_cols`, `class`

```heex
<MyAppBe.MavuDatagrid.datagrid_c
  id={"#{@id}_datagrid"}
  rows={@items_filtered.data}
  visible_cols={[@items_filtered.metadata.columns |> Enum.map(& &1.name)]}
  first_cols="_picker id title"
  last_cols=" _view _edit "
  skip_cols=" created_at"
>
  <:col :let={%{row: row}} for_cols="title">
    <span class="font-bold"><%= row.title %></span>
  </:col>

  <:colheader_style :let={%{col: col}} for_cols="title">
    <span class="text-xs uppercase"><%= col.label %></span>
  </:colheader_style>
</MyAppBe.MavuDatagrid.datagrid_c>
```

---

## MavuList Stack — Filtering, Pagination, Bulk (5/5 projects)

Composed from multiple sub-components. The standard wiring:

| Component | Via | Usage |
|-----------|-----|-------|
| `MavuList.PaginationComponent` | direct | Pagination top/bottom |
| `MavuList.HandleMultipleItemsComponent.buttonbox` | direct | Multi-select bulk actions |
| `MavuList.SearchboxComponent` | direct | Text search |
| `MavuList.ColumnchooserComponent` | direct | Toggle visible columns |
| `MavuList.TagfilterComponent` | direct | Tag-based filtering |
| `MavuList.LabelComponent.paint` | direct | Column labels |
| `MavuListAshfilter.FilterboxLc` | via _bit | Ash-attribute filter form |
| `MavuListAsh.get_query_for_preset!/4` | via _bit | Preset-based query builder |
| `MavuListAsh.get_columns_from_ash_resource/1` | via _bit | Auto-derive columns from Ash resource |

**List component wiring pattern:**
```elixir
def update(assigns, socket) do
  items_query = MavuListAsh.get_query_for_preset!(MyApp.Thing, :default, %{},
    actor: context.current_be_user
  )
  {:ok, socket |> assign(items_query: items_query) |> load_items()}
end

def listconf(assigns) do
  %{
    resource: MyApp.Thing,
    columns: MavuListAsh.get_columns_from_ash_resource(MyApp.Thing)
             |> MavuListAsh.mark_all_hidden_except("id title status"),
    # ...
  }
end

def load_items(socket) do
  socket |> assign(items_filtered: MavuList.process_list(socket.assigns, &listconf/1, &listfilter/3, &default_tweaks/1))
end
```

---

## MavuSubforms — Nested/Embedded Forms (4/5 projects, 56 files)

Manages list-type embedded attributes in AshPhoenix forms.

**Elixir-side — generate subform config for AshPhoenix.Form:**
```elixir
form =
  item
  |> AshPhoenix.Form.for_update(:update,
    [] |> MyAppBe.MavuSubforms.generate_subforms_for_fields(item, tabs.current_subforms)
  )
  |> to_form()
```

**HEEx-side — render subform with add/remove/reorder:**
```heex
<MyAppBe.Bit.MavuSubforms.MavuSubformComponent.main
  field={@form[:skills]}
  phx_target={@myself}
>
  <:itemdiv :let={item_form}>
    <.ash_input field={item_form[:type]} />
    <.ash_input field={item_form[:text]} />
  </:itemdiv>
</MyAppBe.Bit.MavuSubforms.MavuSubformComponent.main>
```

Also uses `MavuSubforms.touch_subform_if_exists/2` (29 calls) to mark subforms as touched during validation.

---

## MavuTags — Tag System (5/5 projects, 86 files)

Provides tag CRUD, resolution, and UI components.

| Function/Component | Calls | Purpose |
|-------------------|-------|---------|
| `MavuTags.Components.tag_label` | 119 | Render a tag pill/badge |
| `MavuTags.handle_event/3` | 31 | Delegate tag-related LiveView events |
| `MavuTags.resolve_uuid/1` | 14 | Resolve tag by UUID or slug |
| `MavuTagsLive.TagfieldC.tagfield_c` | 19 | Tag input field component |

**Tag label in list columns:**
```heex
<:col :let={%{row: row}} for_cols="tags">
  <MyAppBe.MavuTags.Components.tag_label
    :for={tag <- row.tags |> List.wrap()}
    tag={tag}
  />
</:col>
```

---

## MavuBuckets — Key-Value Store (5/5 projects, 61 files)

Server-side key-value storage for ephemeral UI state, editorial positions, and caching.

| Function | Calls | Purpose |
|----------|-------|---------|
| `set_value/4` | 64 | Store a value with optional TTL |
| `get_value/3` | 22 | Retrieve with default |
| `subscribe_live_view/1` | 4 | Subscribe to bucket changes |
| `update_value/4` | 2 | Atomic update with function |

```elixir
# Store
MavuBuckets.set_value("editorial_#{date}", "positions", position_map,
  protect_for: {2, :month}
)

# Retrieve
positions = MavuBuckets.get_value("editorial_#{date}", "positions", %{})

# Atomic update
MavuBuckets.update_value("editorial_#{date}", "additions",
  fn additions -> (additions || %{}) |> Map.put(uuid, value) end,
  %{}
)
```

---

## MavuEntities — Entity Resolution (5/5 projects)

Provides guards, UUID helpers, and cross-entity resolution.

| Function | Calls | Purpose |
|----------|-------|---------|
| `Guards` (import) | 108 | `is_valid_uuid?`, `is_webid?` etc. |
| `UuidHelpers` (alias as `Uuid`) | 102 | `to_webid0/1`, UUID conversion |
| `strip_non_attributes/1` | 34 | Clean struct for serialization |
| `handle_event/3` | 20 | Delegate entity events |

---

## MavuSpec — Resource Metadata (3/5 projects)

Declarative resource config (list presets, column definitions).

```elixir
# Read listconf from resource spec
listconf = MavuSpec.Resource.Info.listconf(MyApp.Event, :default)

# Get column config
col = MavuSpec.Resource.Info.col(MyApp.Event, :title)
```

---

## Extension Points (Delegator Pattern, 5/5 projects)

Projects override shared mappers via `MavuDelegator.delegate_all/1`:

| Mapper | Purpose | Projects |
|--------|---------|----------|
| `MyAppBe.MyAshfilterMapper` | Custom ash filters | 5/5 |
| `MyAppBe.MyCustomInputMapper` | Custom form inputs | 5/5 |
| `MyAppBe.MyBulkeditMapper` | Bulk edit fields | 4/5 |

```elixir
defmodule MyAppBe.MyAshfilterMapper do
  import MavuDelegator

  # project-specific filters here

  delegate_all(MavuBe1.Be.Ashfilters.AshfilterMapper)
end
```

---

## LiveToast — User Feedback (5/5 projects, 46 calls)

```elixir
LiveToast.send_toast(:info, "Saved successfully")
LiveToast.send_toast(:error, "Cannot delete (#{message})")
```

---

## Shorthand `m()` Macro (5/5 projects, 111 files)

Imported via `import Shorthand, warn: false`. Used primarily in `assign_new` pipelines for concise destructuring:

```elixir
|> assign_new(:metrics, fn m(entity) -> get_metrics(entity) end)
# equivalent to:
|> assign_new(:metrics, fn %{entity: entity} -> get_metrics(entity) end)
```
