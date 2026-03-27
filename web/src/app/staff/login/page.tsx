"use client";

import { useState } from "react";
import { useRouter } from "next/navigation";
import { useForm } from "react-hook-form";
import { zodResolver } from "@hookform/resolvers/zod";
import { z } from "zod";
import { setStaffSession } from "@/lib/auth/staff-session";

const STAFF_LOGIN_API_URL = "/api/v1/staff/auth/login";

const loginSchema = z.object({
  email: z.string().email("有効なメールアドレスを入力してください"),
  password: z.string().min(1, "パスワードを入力してください"),
});

type LoginFormData = z.infer<typeof loginSchema>;

interface StaffLoginResponse {
  token: string;
  staff: {
    id: string;
    name: string;
  };
}

interface LoginError {
  error: string;
}

export default function StaffLoginPage() {
  const router = useRouter();
  const [serverError, setServerError] = useState<string | null>(null);

  const {
    register,
    handleSubmit,
    formState: { errors, isSubmitting },
  } = useForm<LoginFormData>({
    resolver: zodResolver(loginSchema),
  });

  const onSubmit = async (data: LoginFormData) => {
    setServerError(null);

    try {
      const response = await fetch(STAFF_LOGIN_API_URL, {
        method: "POST",
        headers: { "Content-Type": "application/json" },
        body: JSON.stringify(data),
      });

      if (!response.ok) {
        const errorData: LoginError = await response.json();
        setServerError(errorData.error || "ログインに失敗しました");
        return;
      }

      const result: StaffLoginResponse = await response.json();
      setStaffSession(result.token);
      router.push("/staff/dashboard");
    } catch {
      setServerError("ネットワークエラーが発生しました。もう一度お試しください。");
    }
  };

  return (
    <div className="min-h-screen flex items-center justify-center bg-gray-50">
      <div className="max-w-md w-full space-y-8 p-8 bg-white rounded-lg shadow">
        <div>
          <h1 className="text-center text-2xl font-bold text-gray-900">
            Progress Checker
          </h1>
          <p className="mt-2 text-center text-sm text-gray-600">
            Staff Login
          </p>
        </div>

        <form className="mt-8 space-y-6" onSubmit={handleSubmit(onSubmit)}>
          {serverError && (
            <div
              className="bg-red-50 border border-red-200 text-red-700 px-4 py-3 rounded"
              role="alert"
            >
              {serverError}
            </div>
          )}

          <div className="space-y-4">
            <div>
              <label
                htmlFor="email"
                className="block text-sm font-medium text-gray-700"
              >
                Email
              </label>
              <input
                id="email"
                type="email"
                autoComplete="email"
                {...register("email")}
                className="mt-1 block w-full px-3 py-2 border border-gray-300 rounded-md shadow-sm focus:outline-none focus:ring-indigo-500 focus:border-indigo-500 sm:text-sm"
                placeholder="staff@example.com"
              />
              {errors.email && (
                <p className="mt-1 text-sm text-red-600">{errors.email.message}</p>
              )}
            </div>

            <div>
              <label
                htmlFor="password"
                className="block text-sm font-medium text-gray-700"
              >
                Password
              </label>
              <input
                id="password"
                type="password"
                autoComplete="current-password"
                {...register("password")}
                className="mt-1 block w-full px-3 py-2 border border-gray-300 rounded-md shadow-sm focus:outline-none focus:ring-indigo-500 focus:border-indigo-500 sm:text-sm"
                placeholder="パスワードを入力"
              />
              {errors.password && (
                <p className="mt-1 text-sm text-red-600">{errors.password.message}</p>
              )}
            </div>
          </div>

          <button
            type="submit"
            disabled={isSubmitting}
            className="w-full flex justify-center py-2 px-4 border border-transparent rounded-md shadow-sm text-sm font-medium text-white bg-indigo-600 hover:bg-indigo-700 focus:outline-none focus:ring-2 focus:ring-offset-2 focus:ring-indigo-500 disabled:opacity-50 disabled:cursor-not-allowed"
          >
            {isSubmitting ? "ログイン中..." : "ログイン"}
          </button>
        </form>
      </div>
    </div>
  );
}
