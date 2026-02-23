import type { User } from "../../types";

interface UserInfoProps {
  user: User;
  onLogout: () => void;
}

export function UserInfo({ user, onLogout }: UserInfoProps) {
  return (
    <div className="flex items-center gap-3">
      <span className="text-sm font-medium text-gray-700">{user.username}</span>
      <button
        onClick={onLogout}
        className="text-xs text-gray-400 hover:text-gray-600"
      >
        Sign out
      </button>
    </div>
  );
}
