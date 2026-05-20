export type User = {
  id: string;
  username: string;
};

export type AuthResponse = {
  user: User;
};

export type LoginRequest = {
  username: string;
  password: string;
};

export type ApiError = {
  error: {
    message: string;
  };
};
