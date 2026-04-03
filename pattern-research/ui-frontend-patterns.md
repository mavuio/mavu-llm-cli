# UI & Frontend Patterns (Elixir-side)

Recurring LiveView, component, and template patterns across 5 Phoenix/Elixir projects.

## Thin screen LiveViews compose list/detail/edit children

**Found in:** 5/5 projects (`wzz-ex`, `climate`, `chatty`, `eselrw`, `filmarchiv-ex`)

**Description:** Top-level `*_live` modules stay thin and mostly decide which child UI module to render. The common shape is: shell LiveView owns routing state, then swaps between a list component and an editor/detail component based on URL params. Directory layout follows the same screen split too: a screen folder usually contains a shell module plus list/detail/edit modules, and 4/5 projects also use numeric filename prefixes like `00_`, `10_`, and `100_` to make render order obvious.

**Example** (from `wzz-ex/lib/my_app_be/products_live/00_products_live.html.ex`):
```elixir
  def render(assigns) do
    ~H"""
    <%= case @live_action do %>
      <% :index -> %>
        <%= if @context.params["product_id"] do %>
          <MyAppBe.RecordPreview.RecordPreviewC.wrapper preview_url={
            if @context.params["product_id"] == "new" do
              nil
            else
              ~p"/embed/#{@context.lang}/shop/product/#{{"p", @context.params["product_id"]} |> Uuid.to_webid0()}?clear_cache=1"
            end
          }>
            <MyAppBe.ProductLive.ProductEditComponent.call
              id="product_edit"
              context={@context}
              rec_id={@context.params["product_id"]}
            />
          </MyAppBe.RecordPreview.RecordPreviewC.wrapper>
        <% else %>
          <MyAppBe.ProductsLive.ProductListLc.call id="productlist" context={@context} />
        <% end %>
    <% end %>
    """
  end
```

**Variations:**
- `climate` and `chatty` often keep the shell even thinner and just render a single list component for `:index`.
- `eselrw` and `filmarchiv-ex` frequently wrap editors in preview/presence UI when a `rec` param is present.

---

## Live components expose a stable `call/1` API

**Found in:** 5/5 projects (`wzz-ex`, `climate`, `chatty`, `eselrw`, `filmarchiv-ex`)

**Description:** Reusable LiveComponents usually publish a small, stable rendering API via `call/1`. The parent calls `SomeLc.call(...)`, while the component itself owns the underlying `<.live_component>` wiring and its typed attrs. This keeps parent templates terse and consistent.

**Example** (from `filmarchiv-ex/lib/my_app_be/tools_live/10_action_runner_lc.html.ex`):
```elixir
defmodule MyAppBe.ToolsLive.ActionRunnerLc do
  use MyAppBe, :live_component

  import MyApp.MavuEntities.Guards, warn: false

  import Shorthand, warn: false

  attr :context, :map, required: true
  attr :tick, :integer, required: true
  attr :id, :string, required: true

  def call(assigns) do
    ~H"""
    <.live_component id={@id} module={__MODULE__} {assigns} />
    """
  end
```

**Variations:**
- List/detail helper components use `call/1` heavily in all 5 projects.
- Some edit components are mounted directly with `<.live_component module={...} ... />` instead of a wrapper, especially when the parent already needs explicit `module=` control.

---

## List pages are built around slot-driven datagrids plus filter/pagination helpers

**Found in:** 5/5 projects (`wzz-ex`, `climate`, `chatty`, `eselrw`, `filmarchiv-ex`)

**Description:** The recurring list-screen stack is: filter component, pagination, optional multi-select/bulk actions, then a `MavuDatagrid.datagrid_c` with HEEx slots for headers, rows, detail rows, and column styling. This is the strongest shared list-template pattern across the projects.

**Example** (from `eselrw/lib/my_app_be/events_live/event_list_component.html.heex`):
```elixir
  <.live_component
    module={MyAppBe.MavuListAshfilter.FilterboxLc}
    id="list_ashfilter"
    list={@items_filtered}
    target={@myself}
    context={@context}
  />

  <.live_component
    module={MyAppBe.MavuList.PaginationComponent}
    list={@items_filtered}
    class="mx-8 mt-8 mb-4"
    id="paginate_list_top"
  />

  <MyAppBe.MavuList.HandleMultipleItemsComponent.buttonbox selected_ids={@selected_ids}>
    <div class="flex space-x-4">
      <.mybutton
        phx-click="update_selected_items"
        phx-target={@myself}
        phx-value-action="delete"
        data-confirm="Are you sure?"
      >
        <.icon name="hero-trash" class="w-5 h-5 mr-4 " />delete
      </.mybutton>

      <.live_component
        module={MyAppBe.Components.BulkEditEntitiesLc}
        id="list_bulkedit"
        list={@items_filtered}
        target={@myself}
        context={@context}
        selected_ids={@selected_ids}
      />
    </div>
  </MyAppBe.MavuList.HandleMultipleItemsComponent.buttonbox>

  <MyAppBe.MavuDatagrid.datagrid_c
    id={"#{@id}_datagrid"}
    rows={@items_filtered.data}
    visible_cols={[@items_filtered.metadata.columns |> Enum.map(& &1.name)]}
    first_cols={
      if @preset not in [:default, :with_legacy],
        do:
          "_picker idv image title_subtitle gehege_tags showtime_stats first_last_day tipps elevation relations ",
        else: "_picker id "
    }
    last_cols=" _view _edit "
    skip_cols=" created_at"
    class="w-full  rounded-lg shadow-sm [&>*>tr>*]:border-x-0"
  >
```

**Variations:**
- `chatty` uses the same datagrid stack in fewer screens, but the same slot vocabulary (`<:col>`, `<:detail_row>`, `<:colheader_style>`) still appears.
- `eselrw`, `climate`, and `filmarchiv-ex` lean hardest on bulk-edit, column-chooser, and ash-filter helpers.

---

## Edit components keep state in `AshPhoenix.Form` and tab-aware event handlers

**Found in:** 5/5 projects (`wzz-ex`, `climate`, `chatty`, `eselrw`, `filmarchiv-ex`)

**Description:** Editor components keep the authoritative UI state in an `AshPhoenix.Form`, then drive it through a repeating event set: `validate`, `save`, `add_form`, `remove_form`, and often `choose_tab`. Tab selection is frequently URL-backed, so moving between tabs and saving are part of the same LiveComponent workflow.

**Example** (from `climate/ui/lib/my_app_be/docjobs_live/docjob_edit_lc.ex`):
```elixir
  def handle_event("validate", %{"form" => params}, socket) do
    form =
      AshPhoenix.Form.validate(socket.assigns.form, params, errors: true)

    {:noreply, assign(socket, form: form)}
  end

  def handle_event("save", %{"form" => params} = msg, socket) do
    options =
      if msg["next_tab"] do
        []
      else
        [params: params |> Map.put("_current_action", "submit")]
      end

    socket =
      case AshPhoenix.Form.submit(socket.assigns.form, options) do
        {:ok, result} ->
          if msg["next_tab"] do
            socket
            |> push_patch(
              to:
                MavuUtils.update_params_in_url(socket.assigns.context.current_url,
                  rec: result.id,
                  tab: msg["next_tab"]
                )
            )
          else
            socket
            |> mymodal_close_via_socket()
            |> push_patch(
              to:
                MavuUtils.update_params_in_url(socket.assigns.context.current_url,
                  rec: nil
                )
            )
          end

        {:error, form} ->
          socket |> assign(form: form)
      end

    {:noreply, socket}
  end

  def handle_event("choose_tab", %{"tab" => tab_key} = _msg, socket) do
    if socket.assigns.form.source.changed? do
      handle_event("save", %{"form" => %{}, "next_tab" => tab_key}, socket)
    else
      {:noreply,
       socket
       |> push_patch(
         to:
           MavuUtils.update_params_in_url(socket.assigns.context.current_url,
             rec: socket.assigns.rec_id,
             tab: tab_key
           )
       )}
    end
  end
```

**Variations:**
- `wzz-ex` includes simpler single-record editors without much tab logic.
- `climate`, `eselrw`, and `filmarchiv-ex` use the tab/subform flow most heavily.
- `chatty` has the same validate/save shape, but often with fewer subforms.

---

## HEEx forms share a common component DSL

**Found in:** 5/5 projects (`wzz-ex`, `climate`, `chatty`, `eselrw`, `filmarchiv-ex`)

**Description:** The form markup layer is extremely consistent: editors use `<.simple_form>` as the wrapper, `<.ash_input>` for fields, `<:actions>` for footer controls, and `<.mybutton>`, `<.icon>`, `<.myheadline>`, and `<.uuid_label>` as the shared UI vocabulary. This is one of the most obvious template-level patterns across all 5 codebases.

**Example** (from `filmarchiv-ex/lib/my_app_be/piles_live/pile_edit_component.html.heex`):
```elixir
  <div role="tablist" class="tabs tabs-box">
    <a
      :for={tab <- @tabs.items}
      role="tab"
      class={Tails.classes(["tab", {"tab-active", tab.active?}])}
      phx-target={@myself}
      phx-click="choose_tab"
      phx-value-tab={tab.key}
    >
      <%= tab.title %>
    </a>
  </div>

 <.simple_form
      :let={_f}
      for={@form}
      as={:form}
      phx-change="validate"
      phx-submit="choose_tab"
      phx-value-tab="close"
      phx-target={@myself}
      phx-mounted={JS.focus(to: "#form_name")}
      id="editform"
      class="space-y-1"
    >
      <%= case @tabs.current do %>
        <% :main-> %>

          <div class="flex gap-4">
            <.ash_input field={@form[:slug]} />
            <.ash_input field={@form[:title]} />
          </div>
      <% end %>
      <button type="submit" class="invisible" title="needed to submit on Enter">OK</button>
      <:actions   :if={save_button_visible?(@tabs.current)}>
        <.mybutton
          phx-disable-with="Sending..."
          type="submit"
          disabled={not @form.source.valid?}
        >
          Save & Close
        </.mybutton>
        <.mybutton navigate={MavuUtils.update_params_in_url(@context.current_url, rec: nil)} >
          Cancel
        </.mybutton>
      </:actions>
    </.simple_form>
```

**Variations:**
- Some editors submit straight to `"save"`; tabbed editors often submit `"choose_tab"` and let the component decide whether to save or just switch tabs.
- `chatty` and `wzz-ex` use the same components in flatter, single-page forms; `eselrw` and `filmarchiv-ex` use them in larger editors with tabs and nested tools.

---

## URL params are the main UI state for opening editors and selecting detail views

**Found in:** 5/5 projects (`wzz-ex`, `climate`, `chatty`, `eselrw`, `filmarchiv-ex`)

**Description:** Instead of hiding state in local assigns only, these projects keep important UI state in the URL: current record id, active tab, selected log/job id, or “new” editor state. That makes list/detail/edit flows patchable, shareable, and back-button-friendly.

**Example** (from `chatty/lib/my_app_be/prompts_live/10_prompts_list_lc.html.ex`):
```elixir
  def handle_event("choose_row", %{"id" => id}, socket) do
    {:noreply,
     socket
     |> push_navigate(
       to:
         MavuUtils.update_params_in_url(socket.assigns.context.current_url,
           prompt_id: if(socket.assigns.context.params["prompt_id"] == id, do: nil, else: id)
         )
     )}
  end
```

**Variations:**
- `wzz-ex`, `climate`, `eselrw`, and `filmarchiv-ex` usually use a generic `rec` param; `chatty` often uses entity-specific params like `prompt_id` or `agent_id`.
- In templates, the same pattern appears as `patch={MavuUtils.update_params_in_url(...)}` or `navigate={MavuUtils.update_params_in_url(...)}` on links and buttons.

---

## Navigation menus are declarative nested trees in `navigation.ex`

**Found in:** 5/5 projects (`wzz-ex`, `climate`, `chatty`, `eselrw`, `filmarchiv-ex`)

**Description:** Backend navigation is defined as plain Elixir data, not embedded directly into templates. Each project builds a `links` list of `%{label, navigate/href, children}` maps and then filters or augments it with role/context rules. This gives the UI a shared menu shape even though the actual sections differ.

**Example** (from `eselrw/lib/my_app_be/navigation.ex`):
```elixir
  def get_menu(context) when is_map(context) do
    links = [
      %{label: "home", navigate: ~p"/be"},
      %{
        label: "Agents",
        navigate: ~p"/be/agents",
        allowed_roles: [:admin]
      },
      %{
        label: "Albums",
        navigate: ~p"/be/albums",
        children: [
          %{label: "All", navigate: ~p"/be/albums"},
          %{label: "From Flickr", navigate: ~p"/be/albums?preset=from_flickr"},
          %{
            label: "Create Album from Flickr",
            navigate: ~p"/be/tools/create_album_from_flickr"
          }
        ],
        allowed_roles: [:admin]
      },
      %{
        label: "Events",
        children: [
          %{label: "Daytool", navigate: ~p"/be/tools/daytool", allowed_roles: [:admin]},
          %{label: "Weektool", navigate: ~p"/be/tools/weektool", allowed_roles: [:admin]},
          %{label: "Piles", navigate: ~p"/be/piles", allowed_roles: [:admin]}
        ],
        allowed_roles: [:admin, :guest]
      }
    ]
```

**Variations:**
- `climate` leans more on `href` than `navigate` and builds some menu items from bookmark data.
- `chatty`, `eselrw`, and `wzz-ex` mix static children with bookmark-derived children.
- `filmarchiv-ex` and `eselrw` have the largest nested admin menus.

---

## Function components use typed attrs/slots as reusable UI APIs

**Found in:** 5/5 projects (`wzz-ex`, `climate`, `chatty`, `eselrw`, `filmarchiv-ex`)

**Description:** Beyond LiveComponents, all 5 projects also use function components with typed `attr` declarations, `slot` APIs, `assign_new/3`, and `render_slot/2`. This is the main pattern for reusable layout wrappers and small view-level abstractions.

**Example** (from `chatty/lib/my_app_be/lazy_tab_wrapper_c.html.ex`):
```elixir
  attr :tabs, :list, required: true, doc: "the tabs to render, needs a label and a slug parameter"

  attr :filtervar_prefix, :string, default: "tab_"
  attr :class, :any, default: [], doc: "classes"

  attr :stylekey,
       :atom,
       default: :lifted,
       doc: "stylekey: :lifted, :bordered, :boxed, :pill",
       values: [:lifted, :bordered, :boxed, :pill]

  attr :h_align, :any,
    default: :start,
    doc: "horizontal alignment: :start, :center, :end",
    values: [:start, :end]

  attr :context, :any, required: true
  attr :rest, :global

  slot :inner_block, required: true

  slot :slot_a, required: false

  def lazy_tab_wrapper(assigns) do
    assigns =
      assigns
      |> assign_new(:filter_vars, fn m(context, filtervar_prefix) ->
        MyAppBe.Components.Filters.FilterHelpers.get_filter_vars(%{
          filtervar_prefix: filtervar_prefix,
          context: context,
          json_config: %{}
        })
      end)
```

**Variations:**
- The exact helper names differ per project.
- `lazy_tab_wrapper_c` specifically shows up in `chatty` and `filmarchiv-ex`, while the same attr/slot style appears in smaller wrappers and custom components in the other projects.

---
