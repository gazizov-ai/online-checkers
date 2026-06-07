import { UserRound } from "lucide-react";
import { useEffect, useState } from "react";
import type { Profile } from "../types/domain";
import { profileName } from "../api/profile";

type ProfileAvatarProps = {
  profile: Profile;
  size?: "small" | "large";
};

export function ProfileAvatar({ profile, size = "small" }: ProfileAvatarProps) {
  const [imageFailed, setImageFailed] = useState(false);
  const name = profileName(profile);

  useEffect(() => {
    setImageFailed(false);
  }, [profile.avatar_url]);

  return (
    <span className={`profile-avatar profile-avatar--${size}`} title={name}>
      {profile.avatar_url && !imageFailed ? (
        <img
          alt=""
          onError={() => setImageFailed(true)}
          src={profile.avatar_url}
        />
      ) : size === "large" ? (
        <span aria-hidden>{name.slice(0, 1).toLocaleUpperCase("ru")}</span>
      ) : (
        <UserRound aria-hidden size={18} />
      )}
    </span>
  );
}
