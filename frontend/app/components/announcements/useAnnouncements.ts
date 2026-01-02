import { useCallback, useEffect, useMemo, useState } from "react";

import { AnnouncementService, type Announcement } from "@/services/admin/adminService";
import { useMutation, useQuery, useQueryClient } from "@tanstack/react-query";

const DISMISSED_KEY = "dismissed_announcements";

function getDismissedIds(): Set<number> {
  try {
    const stored = localStorage.getItem(DISMISSED_KEY);
    if (stored) {
      return new Set(JSON.parse(stored));
    }
  } catch {
    // Ignore parse errors
  }
  return new Set();
}

function saveDismissedIds(ids: Set<number>) {
  localStorage.setItem(DISMISSED_KEY, JSON.stringify([...ids]));
}

export function useAnnouncements() {
  const queryClient = useQueryClient();
  const [dismissedIds, setDismissedIds] = useState<Set<number>>(() => getDismissedIds());
  const [currentModalIndex, setCurrentModalIndex] = useState(0);
  const [modalQueue, setModalQueue] = useState<Announcement[]>([]);
  const [isModalOpen, setIsModalOpen] = useState(false);

  // Fetch active banner announcements
  const {
    data: allAnnouncements = [],
    isLoading: isLoadingBanners,
    error: bannersError,
  } = useQuery({
    queryKey: ["announcements", "active"],
    queryFn: AnnouncementService.getActiveAnnouncements,
    staleTime: 5 * 60 * 1000, // 5 minutes
    refetchOnWindowFocus: false,
  });

  // Fetch unread modal announcements (for authenticated users)
  const { data: unreadModals = [], isLoading: isLoadingModals } = useQuery({
    queryKey: ["announcements", "unread-modals"],
    queryFn: AnnouncementService.getUnreadModalAnnouncements,
    staleTime: 5 * 60 * 1000,
    refetchOnWindowFocus: false,
    retry: false, // Don't retry if user is not authenticated
  });

  // Filter banner announcements (not dismissed, display_type = banner)
  const bannerAnnouncements = useMemo(() => {
    return allAnnouncements.filter((a) => a.display_type === "banner" && !dismissedIds.has(a.id));
  }, [allAnnouncements, dismissedIds]);

  // Set up modal queue when unread modals are fetched
  useEffect(() => {
    if (unreadModals.length > 0 && modalQueue.length === 0) {
      setModalQueue(unreadModals);
      setCurrentModalIndex(0);
      setIsModalOpen(true);
    }
  }, [unreadModals, modalQueue.length]);

  // Dismiss banner mutation
  const { mutate: dismissMutate, isPending: isDismissing } = useMutation({
    mutationFn: AnnouncementService.dismissAnnouncement,
    onSuccess: (_, announcementId) => {
      // Update local state
      setDismissedIds((prev) => {
        const next = new Set(prev);
        next.add(announcementId);
        saveDismissedIds(next);
        return next;
      });
    },
  });

  // Mark modal as read mutation
  const { mutate: markReadMutate } = useMutation({
    mutationFn: AnnouncementService.markAnnouncementRead,
    onSuccess: () => {
      // Invalidate the unread modals query
      void queryClient.invalidateQueries({
        queryKey: ["announcements", "unread-modals"],
      });
    },
  });

  const dismissBanner = useCallback(
    (id: number) => {
      dismissMutate(id);
    },
    [dismissMutate]
  );

  const markModalRead = useCallback(
    (id: number) => {
      markReadMutate(id);
    },
    [markReadMutate]
  );

  const closeCurrentModal = useCallback(() => {
    if (currentModalIndex < modalQueue.length - 1) {
      // Move to next modal
      setCurrentModalIndex((prev) => prev + 1);
    } else {
      // No more modals
      setIsModalOpen(false);
      setModalQueue([]);
      setCurrentModalIndex(0);
    }
  }, [currentModalIndex, modalQueue.length]);

  const currentModal = modalQueue[currentModalIndex] || null;

  return {
    // Banner announcements
    bannerAnnouncements,
    isLoadingBanners,
    bannersError,
    dismissBanner,
    isDismissing,

    // Modal announcements
    currentModal,
    isModalOpen,
    closeCurrentModal,
    markModalRead,
    isLoadingModals,
    modalCount: modalQueue.length,
    currentModalNumber: currentModalIndex + 1,
  };
}
