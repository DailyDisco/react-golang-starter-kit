import { AnnouncementBanner } from "./AnnouncementBanner";
import { AnnouncementModal } from "./AnnouncementModal";
import { useAnnouncements } from "./useAnnouncements";

export function AnnouncementsContainer() {
  const { bannerAnnouncements, dismissBanner, currentModal, isModalOpen, closeCurrentModal, markModalRead } =
    useAnnouncements();

  return (
    <>
      {/* Banner announcements */}
      {bannerAnnouncements.map((announcement) => (
        <AnnouncementBanner
          key={announcement.id}
          announcement={announcement}
          onDismiss={dismissBanner}
        />
      ))}

      {/* Modal announcement */}
      <AnnouncementModal
        announcement={currentModal}
        open={isModalOpen}
        onClose={closeCurrentModal}
        onMarkRead={markModalRead}
      />
    </>
  );
}
