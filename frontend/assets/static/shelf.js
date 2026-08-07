(function () {
  // Pointer-driven shelf reordering. Native drag/drop is avoided because shelf
  // cards contain buttons/popovers, which makes browser drop targeting brittle.
  const dragThreshold = 6;
  const scrollEdgeSize = 80;
  const maxScrollSpeed = 22;

  // Pointer state tracks both drag start and current viewport coordinates.
  let pointerID = null;
  let startX = 0;
  let startY = 0;
  let lastPointerX = 0;
  let lastPointerY = 0;

  // Drag state: pending preserves click behavior until movement crosses the threshold.
  let pendingSlot = null;
  let draggedSlot = null;
  let hoverSlot = null;

  // Preview state stores each slot's original CSS order so swaps can be shown
  // without moving DOM nodes or disturbing popover markup.
  let slots = [];
  let originalOrders = new Map();

  // Side-effect state for async persistence, the floating card, and edge scrolling.
  let dropInProgress = false;
  let dragGhost = null;
  let ghostOffsetX = 0;
  let ghostOffsetY = 0;
  let scrollFrame = null;

  function shelfSlotFromTarget(target) {
    return target instanceof Element ? target.closest("[data-shelf-card]") : null;
  }

  // Hit-test the occupied shelf slot under a viewport point. The floating ghost
  // has pointer-events disabled, so it will not hide the real shelf target.
  function shelfSlotFromPoint(x, y) {
    return shelfSlotFromTarget(document.elementFromPoint(x, y));
  }

  // Used to distinguish gaps inside the shelf from leaving the shelf entirely.
  function isPointInsideShelf(x, y) {
    const target = document.elementFromPoint(x, y);
    return target instanceof Element && Boolean(target.closest(".shelf-stage, .shelf-grid"));
  }

  // Do not start a drag from controls that need normal click/form behavior.
  function isInteractiveTarget(target) {
    if (!(target instanceof Element)) {
      return false;
    }
    return Boolean(target.closest(".shelf-popover, form, a, input, select, textarea, .popover-close"));
  }

  function clearPreview() {
    document.querySelectorAll("[data-shelf-card].is-drag-preview, [data-shelf-card].is-drop-target").forEach((slot) => {
      slot.classList.remove("is-drag-preview", "is-drop-target");
    });
  }

  // Assign explicit CSS order values before previewing. Swapping order values is
  // much safer than physically moving nodes while the pointer is active.
  function snapshotOrder(startSlot) {
    const grid = startSlot.closest(".shelf-grid");
    slots = grid ? Array.from(grid.querySelectorAll("[data-shelf-card]")) : [];
    originalOrders = new Map();
    slots.forEach((slot, index) => {
      const order = String(index + 1);
      originalOrders.set(slot, order);
      slot.style.order = order;
    });
  }

  // Restore the previewed layout back to the snapshotted order.
  function restoreOrder() {
    originalOrders.forEach((order, slot) => {
      slot.style.order = order;
    });
  }

  // After a successful server refresh, remove temporary order values from old nodes.
  function clearInlineOrder() {
    slots.forEach((slot) => {
      slot.style.order = "";
    });
  }

  function removeDragGhost() {
    if (dragGhost) {
      dragGhost.remove();
    }
    dragGhost = null;
    ghostOffsetX = 0;
    ghostOffsetY = 0;
  }

  // The ghost keeps the dragged card visually under the pointer while the real
  // shelf slot stays in-grid for the low-opacity swap preview.
  function moveDragGhost(x, y) {
    if (!dragGhost) {
      return;
    }
    dragGhost.style.transform = `translate3d(${x - ghostOffsetX}px, ${y - ghostOffsetY}px, 0)`;
  }

  function createDragGhost(slot, event) {
    removeDragGhost();
    const card = slot.querySelector(".shelf-card");
    if (!card) {
      return;
    }
    const rect = slot.getBoundingClientRect();
    ghostOffsetX = event.clientX - rect.left;
    ghostOffsetY = event.clientY - rect.top;
    dragGhost = document.createElement("div");
    dragGhost.className = "shelf-drag-ghost";
    dragGhost.style.width = `${rect.width}px`;
    dragGhost.appendChild(card.cloneNode(true));
    document.body.appendChild(dragGhost);
    moveDragGhost(event.clientX, event.clientY);
  }

  // Re-evaluate the target under a stationary pointer while auto-scroll moves the page.
  function previewAtPointer() {
    if (!draggedSlot) {
      return;
    }
    const targetSlot = shelfSlotFromPoint(lastPointerX, lastPointerY);
    if (targetSlot && targetSlot !== draggedSlot) {
      showPreview(targetSlot);
      return;
    }
    clearHoverIfEmpty(targetSlot, lastPointerX, lastPointerY);
  }

  function scrollSpeedForPointer() {
    if (lastPointerY < scrollEdgeSize) {
      return -maxScrollSpeed * Math.min((scrollEdgeSize - lastPointerY) / scrollEdgeSize, 1);
    }
    const bottomDistance = window.innerHeight - lastPointerY;
    if (bottomDistance < scrollEdgeSize) {
      return maxScrollSpeed * Math.min((scrollEdgeSize - bottomDistance) / scrollEdgeSize, 1);
    }
    return 0;
  }

  // Scroll loop runs only during an active drag and ramps speed near viewport edges.
  function stopAutoScroll() {
    if (scrollFrame !== null) {
      cancelAnimationFrame(scrollFrame);
    }
    scrollFrame = null;
  }

  function autoScroll() {
    if (!draggedSlot || dropInProgress) {
      stopAutoScroll();
      return;
    }
    const speed = scrollSpeedForPointer();
    if (speed !== 0) {
      const before = window.scrollY;
      window.scrollBy(0, speed);
      // At scroll boundaries, avoid recalculating preview against an unchanged viewport.
      if (window.scrollY === before) {
        scrollFrame = requestAnimationFrame(autoScroll);
        return;
      }
      moveDragGhost(lastPointerX, lastPointerY);
      previewAtPointer();
    }
    scrollFrame = requestAnimationFrame(autoScroll);
  }

  function startAutoScroll() {
    stopAutoScroll();
    scrollFrame = requestAnimationFrame(autoScroll);
  }

  // Show the resulting swap by exchanging the dragged and hovered CSS order values.
  function showPreview(targetSlot) {
    if (!draggedSlot || !targetSlot || targetSlot === draggedSlot || targetSlot === hoverSlot) {
      return;
    }
    restoreOrder();
    clearPreview();
    hoverSlot = targetSlot;
    draggedSlot.style.order = originalOrders.get(targetSlot) || "";
    targetSlot.style.order = originalOrders.get(draggedSlot) || "";
    draggedSlot.classList.add("is-drag-preview");
    targetSlot.classList.add("is-drag-preview", "is-drop-target");
  }

  // Preserve the last valid hover while crossing row gaps inside the shelf. Resetting
  // order during auto-scroll can trigger scroll anchoring jumps in some browsers.
  function clearHoverIfEmpty(targetSlot, x, y) {
    if (!draggedSlot || targetSlot === hoverSlot) {
      return;
    }
    if (targetSlot === draggedSlot) {
      return;
    }
    if (!targetSlot && !isPointInsideShelf(x, y)) {
      restoreOrder();
      clearPreview();
      hoverSlot = null;
      draggedSlot.classList.add("is-drag-preview");
    }
  }

  function resetDragState() {
    stopAutoScroll();
    removeDragGhost();
    document.body.classList.remove("is-shelf-dragging");
    clearPreview();
    if (draggedSlot) {
      draggedSlot.classList.remove("is-dragging");
    }
    if (pendingSlot) {
      pendingSlot.classList.remove("is-drag-pending");
    }
    pointerID = null;
    pendingSlot = null;
    draggedSlot = null;
    hoverSlot = null;
    slots = [];
    originalOrders = new Map();
    dropInProgress = false;
    lastPointerX = 0;
    lastPointerY = 0;
  }

  // Persist through the same-origin frontend action so the browser never calls
  // the backend API directly and receives a fresh server-rendered shelf shell.
  async function persistSwap(sourceID, targetID) {
    const body = new URLSearchParams({
      figurine_id: sourceID,
      target_figurine_id: targetID,
      next: "/",
    });
    const response = await fetch("/actions/shelf/swap", {
      method: "POST",
      headers: {
        "Content-Type": "application/x-www-form-urlencoded",
        "HX-Request": "true",
      },
      body,
    });
    if (!response.ok) {
      throw new Error("Shelf swap failed");
    }
    return response.text();
  }

  // Swap the page shell with canonical server HTML after persistence succeeds.
  function replacePageShell(html) {
    const parser = new DOMParser();
    const doc = parser.parseFromString(html, "text/html");
    const nextShell = doc.querySelector("main.page-shell");
    const currentShell = document.querySelector("main.page-shell");
    if (nextShell && currentShell) {
      currentShell.replaceWith(nextShell);
      if (window.htmx) {
        window.htmx.process(nextShell);
      }
    }
  }

  // A drag starts only after pointer movement crosses dragThreshold, preserving
  // normal click-to-open behavior for shelf card detail popovers.
  function startDrag(slot, event) {
    draggedSlot = slot;
    pendingSlot = null;
    hoverSlot = null;
    dropInProgress = false;
    snapshotOrder(slot);
    createDragGhost(slot, event);
    startAutoScroll();
    document.body.classList.add("is-shelf-dragging");
    slot.classList.remove("is-drag-pending");
    slot.classList.add("is-dragging", "is-drag-preview");
    if (slot.setPointerCapture) {
      slot.setPointerCapture(event.pointerId);
    }
  }

  document.addEventListener("pointerdown", (event) => {
    if (event.button !== 0 || isInteractiveTarget(event.target)) {
      return;
    }
    const slot = shelfSlotFromTarget(event.target);
    if (!slot || !slot.dataset.figurineId) {
      return;
    }
    pointerID = event.pointerId;
    pendingSlot = slot;
    startX = event.clientX;
    startY = event.clientY;
    lastPointerX = event.clientX;
    lastPointerY = event.clientY;
    slot.classList.add("is-drag-pending");
  });

  document.addEventListener("pointermove", (event) => {
    if (event.pointerId !== pointerID || dropInProgress) {
      return;
    }
    lastPointerX = event.clientX;
    lastPointerY = event.clientY;
    if (pendingSlot && !draggedSlot) {
      // Movement under the threshold remains a normal click on the shelf card button.
      const distance = Math.hypot(event.clientX - startX, event.clientY - startY);
      if (distance < dragThreshold) {
        return;
      }
      event.preventDefault();
      startDrag(pendingSlot, event);
    }
    if (!draggedSlot) {
      return;
    }
    event.preventDefault();
    moveDragGhost(event.clientX, event.clientY);
    previewAtPointer();
  });

  // Use the last valid hover first: after CSS-order preview, elementFromPoint can
  // report the dragged slot even though the user is visually over the swap target.
  document.addEventListener("pointerup", async (event) => {
    if (event.pointerId !== pointerID) {
      return;
    }
    if (pendingSlot && !draggedSlot) {
      pendingSlot.classList.remove("is-drag-pending");
      resetDragState();
      return;
    }
    if (!draggedSlot) {
      resetDragState();
      return;
    }
    event.preventDefault();
    const targetSlot = hoverSlot || shelfSlotFromPoint(event.clientX, event.clientY);
    if (!targetSlot || targetSlot === draggedSlot) {
      restoreOrder();
      resetDragState();
      return;
    }
    const sourceID = draggedSlot.dataset.figurineId;
    const targetID = targetSlot.dataset.figurineId;
    if (!sourceID || !targetID) {
      restoreOrder();
      resetDragState();
      return;
    }
    dropInProgress = true;
    try {
      const html = await persistSwap(sourceID, targetID);
      clearInlineOrder();
      resetDragState();
      replacePageShell(html);
    } catch (_error) {
      restoreOrder();
      resetDragState();
    }
  });

  document.addEventListener("pointercancel", (event) => {
    if (event.pointerId !== pointerID) {
      return;
    }
    restoreOrder();
    resetDragState();
  });
})();
