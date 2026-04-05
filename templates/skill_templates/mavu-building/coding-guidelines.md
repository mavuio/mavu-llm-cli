# Mavu Phoenix Coding Guidelines

Compact reference for coding agents building backend UI in Mavu Phoenix projects.

---

## File & Module Naming

- **Suffixes encode role:** `*Live` = top-level screen, `*Lc` = live component, `*C` = function component
- **Numeric filename prefixes** set render/reading order within a screen folder: `00_`, `10_`, `20_`, `100_`
- **Screen folders** are named `{entity_plural}_live/` and contain shell + list + detail + edit modules
- **Extension-point dirs** live at `lib/my_app_be/`: `_components/`, `_controllers/`, `my_ashfilters/`, `my_custom_inputs/`, `my_bulkedit_components/`

## Module Boilerplate

Standard imports for UI modules:

```elixir
use MyAppBe, :live_view        # screen shells
use MyAppBe, :live_component   # stateful components
use MyAppBe, :html             # function components

import Shorthand, warn: false
import MyApp.MavuEntities.Guards, warn: false
alias MyApp.MavuEntities.UuidHelpers, as: Uuid
```

Only add imports you actually use. Simpler modules can skip helper imports.

## Screen Shell (LiveView)

Keep shells thin. They own routing state and delegate rendering to child components.

```elixir
defmodule MyAppBe.ThingsLive do
  use MyAppBe, :live_view

  @impl true
  def mount(_params, _session, socket), do: {:ok, socket}

  @impl true
  def handle_params(params, _url, socket) do
    {:noreply, apply_action(socket, socket.assigns.live_action, params)}
  end

  defp apply_action(socket, :index, params) do
    # switch between list and editor based on URL params
  end
end
```

In `render/1`, branch on `@live_action` and URL params to show list vs detail/edit child components.

## Live Components with `call/1`

Expose a `call/1` wrapper so parents render components with a stable API:

```elixir
defmodule MyAppBe.ThingDetailLc do
  use MyAppBe, :live_component

  attr :context, :map, required: true
  attr :id, :string, required: true

  def call(assigns) do
    ~H"""
    <.live_component id={@id} module={__MODULE__} {assigns} />
    """
  end

  # update/2 and render/1 below...
end
```

Parent usage: `<MyAppBe.ThingDetailLc.call id="detail" context={@context} />`

## Function Components — Typed Attrs & Slots

Always declare `attr` and `slot` for function components:

```elixir
attr :items, :list, required: true
attr :class, :any, default: []
attr :rest, :global
slot :inner_block, required: true

def my_widget(assigns) do
  ~H"""
  <div class={@class} {@rest}>
    <%= render_slot(@inner_block) %>
  </div>
  """
end
```

## Derived State via Chained `assign_new`

Build derived assigns as a pipeline. Each step can reference assigns from previous steps:

```elixir
def my_component(assigns) do
  assigns =
    assigns
    |> assign_new(:record, fn %{record_id: id} ->
      MyApp.Thing.resolve!(id)
    end)
    |> assign_new(:is_active?, fn m(record) ->
      record.status == :active
    end)
    |> assign_new(:display_class, fn %{is_active?: active?} ->
      if active?, do: "text-green-600", else: "text-gray-400"
    end)

  ~H"..."
end
```

Three conventions exist — use whichever fits:
- `fn %{key: val} ->` — explicit destructuring
- `fn m(key1, key2) ->` — `Shorthand.m/1` macro
- `fn _ -> default end` — fallback/default values

## List Screens

Standard stack: filter → pagination → optional bulk actions → datagrid.

```elixir
# In update/2:
items_query = MavuListAsh.get_query_for_preset!(MyApp.Thing, preset_atom, %{},
  actor: context.current_be_user
)
socket |> assign(items_query: items_query) |> load_items()
```

Define columns and tweaks in `listconf/1` and `default_tweaks/1`.

In HEEx, compose:
- `FilterboxLc` for ash filters
- `PaginationComponent` top and bottom
- `HandleMultipleItemsComponent.buttonbox` for bulk actions
- `MavuDatagrid.datagrid_c` with `<:col>`, `<:detail_row>`, `<:colheader_style>` slots

## Edit Components (AshPhoenix Forms)

Standard event set: `validate`, `save`, `add_form`, `remove_form`, `choose_tab`.

```elixir
def handle_event("validate", %{"form" => params}, socket) do
  form = AshPhoenix.Form.validate(socket.assigns.form, params, errors: true)
  {:noreply, assign(socket, form: form)}
end

def handle_event("save", %{"form" => params} = msg, socket) do
  case AshPhoenix.Form.submit(socket.assigns.form, params: params) do
    {:ok, result} ->
      {:noreply, socket |> push_patch(to: close_url(socket))}
    {:error, form} ->
      {:noreply, assign(socket, form: form)}
  end
end

def handle_event("choose_tab", %{"tab" => tab_key}, socket) do
  if socket.assigns.form.source.changed? do
    handle_event("save", %{"form" => %{}, "next_tab" => tab_key}, socket)
  else
    {:noreply, socket |> push_patch(to: tab_url(socket, tab_key))}
  end
end
```

In `update/2`, branch on `new` vs existing record, build `AshPhoenix.Form`, store `@tabconf`.

## HEEx Form DSL

Use the shared component vocabulary:

```heex
<.simple_form for={@form} as={:form} phx-change="validate" phx-submit="save" phx-target={@myself}>
  <.ash_input field={@form[:name]} />
  <.ash_input field={@form[:email]} />

  <:actions>
    <.mybutton type="submit" disabled={not @form.source.valid?}>Save</.mybutton>
    <.mybutton navigate={cancel_url(@context)}>Cancel</.mybutton>
  </:actions>
</.simple_form>
```

For tabbed forms, submit to `"choose_tab"` with `phx-value-tab="close"` instead of `"save"`.

Tabs use DaisyUI:
```heex
<div role="tablist" class="tabs tabs-box">
  <a :for={tab <- @tabs.items}
     role="tab"
     class={Tails.classes(["tab", {"tab-active", tab.active?}])}
     phx-target={@myself} phx-click="choose_tab" phx-value-tab={tab.key}>
    <%= tab.title %>
  </a>
</div>
```

## URL-Driven UI State

Keep editor/detail state in URL params — not hidden in assigns:

```elixir
# Opening an editor
push_patch(to: MavuUtils.update_params_in_url(socket.assigns.context.current_url, rec: id))

# Toggling a detail row
push_navigate(to: MavuUtils.update_params_in_url(url, rec: toggle(current, id)))

# Closing
push_patch(to: MavuUtils.update_params_in_url(url, rec: nil))
```

Prefer the generic `rec` param for the current record and `tab` for the active tab.

## Result Handling

Branch on `{:ok, result}` / `{:error, form_or_changeset}` tuples. Use `LiveToast` for user feedback:

```elixir
case Ash.destroy(item) do
  :ok ->
    LiveToast.send_toast(:info, "Deleted")
    {:noreply, push_navigate(socket, to: list_url)}
  {:error, %{errors: [%{message: msg}]}} ->
    LiveToast.send_toast(:error, "Cannot delete (#{msg})")
    {:noreply, socket}
end
```

For form saves, reassign the error form instead of toasting.

## Navigation

Define in `navigation.ex` as a plain data tree:

```elixir
def get_menu(context) do
  [
    %{label: "Home", navigate: ~p"/be"},
    %{label: "Things", navigate: ~p"/be/things", allowed_roles: [:admin]},
    %{label: "Tools", children: [
      %{label: "Import", navigate: ~p"/be/tools/import"},
      %{label: "Export", navigate: ~p"/be/tools/export"}
    ]}
  ]
end
```

## Extension Points (Delegator Pattern)

Local overrides delegate to shared base mappers:

```elixir
defmodule MyAppBe.MyAshfilterMapper do
  import MavuDelegator
  # local filters go here
  delegate_all(MavuBe1.Be.Ashfilters.AshfilterMapper)
end
```

Same pattern for `MyCustomInputMapper` and `MyBulkeditMapper`.
